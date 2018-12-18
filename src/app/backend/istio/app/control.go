// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"errors"
	"fmt"
	"log"
	"time"

	api2 "github.com/kubernetes/dashboard/src/app/backend/api"
	kdErrors "github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/destinationrule"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CanaryApp creates a canary version for the specified namespace
// version is a logic canary meaning, doesn't need to bind to image version.
func CanaryApp(client kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery,
	appName string, canaryDep *api.CanaryDeployment) error {
	version := canaryDep.Version

	// check if the specified app exist
	dataSelector := dataselect.NoDataSelect
	dataSelector.FilterQuery = dataselect.NewFilterQuery([]string{dataselect.NameProperty, appName})

	_, err := GetAppDetail(client, istioClient, namespace, appName, dataSelector)
	if err != nil {
		return err
	}

	// check if the app is in canary
	// if multiple destination rules exist
	dRules, err := destinationrule.GetDestinationRuleListByHostname(client, istioClient, namespace, []string{appName})
	if err != nil {
		return err
	}

	if dRules.ListMeta.TotalItems > 1 {
		return fmt.Errorf("app %s is in canary", appName)
	}

	// 3. create a deployment with specified version & canary plan name
	// find the existed deployment first, and inherent from its deployment configuration
	var parent v1beta1.Deployment
	if parentDeps, err := getDeploymentByLabels(client, namespace, map[string]string{"app": appName}); err != nil || len(parentDeps) != 1 {
		return fmt.Errorf("support only one parent deployment, %d given", len(parentDeps))
	} else {
		parent = parentDeps[0]
	}

	// fix parent deployment when some label is not set correctly.
	if err := fixDeployment(client, &parent); err != nil {
		return err
	}

	// create new deployment
	newPodSpec := canaryDep.PodTemplate
	newPodSpec.Labels["app"] = appName
	newPodSpec.Labels["qcloud-app"] = appName
	newPodSpec.Labels["version"] = version

	var replica int32
	if canaryDep.Replicas > 0 {
		replica = canaryDep.Replicas
	} else {
		replica = *parent.Spec.Replicas
	}
	newDep := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       api2.ResourceKindDeployment,
			APIVersion: parent.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", appName, version),
			Namespace: parent.Namespace,
			Labels: map[string]string{
				"app":        appName,
				"qcloud-app": appName,
				"version":    version,
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     appName,
					"version": version,
				},
			},
			Template:                newPodSpec,
			Strategy:                parent.Spec.Strategy,
			MinReadySeconds:         parent.Spec.MinReadySeconds,
			RevisionHistoryLimit:    parent.Spec.RevisionHistoryLimit,
			ProgressDeadlineSeconds: parent.Spec.ProgressDeadlineSeconds,
		},
	}

	newDep, err = client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).Create(newDep)
	if err != nil {
		return err
	}

	// create destination rules
	destinationRule, err := istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Get(appName, metav1.GetOptions{})
	if err != nil {
		if kdErrors.IsNotFoundError(err) {
			// create a new destinationRule
			rule := &istioApi.DestinationRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: namespace.ToRequestParam(),
				},
				Spec: istioApi.DestinationRuleSpec{
					Host: appName,
					Subsets: []*istioApi.Subset{
						{
							Name: parent.Labels["version"],
							Labels: map[string]string{
								"version": parent.Labels["version"],
							},
						},
						{
							Name: newDep.Labels["version"],
							Labels: map[string]string{
								"version": newDep.Labels["version"],
							},
						},
					},
				},
			}
			_, err := istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Create(rule)
			return err
		}
		return err
	}

	return addToDestinationRule(istioClient, destinationRule, version, namespace.ToRequestParam())
}

// CreateApp creates application
// 1. create service
// 2. create deployment, app, labels, podTemplate and so on.
// 3. create destination rule
func CreateApp(client kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery, appName string, newApp *api.NewApplication) error {
	// TODO validation
	var err error
	version := newApp.Version
	if version == "" {
		return errors.New("Application creation without app version")
	}

	// 1. Create service
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       api2.ResourceKindService,
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace.ToRequestParam(),
			Labels: map[string]string{
				"app":        appName,
				"qcloud-app": appName,
			},
		},
		Spec: v1.ServiceSpec{
			Ports: newApp.Ports,
			Selector: map[string]string{
				"app": appName,
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
	_, err = client.CoreV1().Services(namespace.ToRequestParam()).Create(svc)
	if err != nil {
		return err
	}

	var replica int32
	if newApp.Replicas > 0 {
		replica = newApp.Replicas
	} else {
		// default value
		replica = 2
	}

	newPodSpec := newApp.PodTemplate
	newPodSpec.Labels["app"] = appName
	newPodSpec.Labels["qcloud-app"] = appName
	newPodSpec.Labels["version"] = version

	var limit int32 = 5
	var deadlineSeconds int32 = 600
	newDep := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       api2.ResourceKindDeployment,
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", appName, version),
			Namespace: namespace.ToRequestParam(),
			Labels: map[string]string{
				"app":        appName,
				"qcloud-app": appName,
				"version":    version,
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     appName,
					"version": version,
				},
			},
			Template: newPodSpec,
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
			},
			MinReadySeconds:         10,
			RevisionHistoryLimit:    &limit,
			ProgressDeadlineSeconds: &deadlineSeconds,
		},
	}

	_, err = client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).Create(newDep)
	if err != nil {
		return err
	}

	// 3. Create destination rule
	rule := &istioApi.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace.ToRequestParam(),
		},
		Spec: istioApi.DestinationRuleSpec{
			Host: appName,
			Subsets: []*istioApi.Subset{
				{
					Name: version,
					Labels: map[string]string{
						"version": version,
					},
				},
			},
		},
	}
	_, err = istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Create(rule)
	if err != nil {
		return err
	}
	return nil
}

func OfflineAppVersion(client kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery,
	appName string, version string, offlineType string) error {
	var (
		virtualServices []istioApi.VirtualService
		err             error
	)

	virtualServices, err = virtualservice.GetVirtualServices(istioClient, []string{appName}, offlineType)
	if err != nil {
		return err
	}

	for _, vs := range virtualServices {
		err = removeFromVirtualService(istioClient, &vs, version)
		if err != nil {
			log.Println("fail to remove from virtual service", err)
		}
	}

	dRule, err := istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Get(appName, metav1.GetOptions{})
	if err != nil {
		log.Println("fail to get destination rule", err)
	}

	if err = removeFromDestinationRule(istioClient, dRule, version); err != nil {
		log.Println("fail to remove from destination rule", err)
	}

	// Sleep 3 seconds for changing the flow
	time.Sleep(3 * time.Second)

	deps, err := getDeploymentByLabels(client, namespace, map[string]string{
		"app":     appName,
		"version": version,
	})
	if err != nil {
		return err
	}

	if len(deps) != 1 {
		return fmt.Errorf("deployment not found, %d given", len(deps))
	}

	replicaDeletion := metav1.DeletePropagationBackground
	err = client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).Delete(
		deps[0].Name, &metav1.DeleteOptions{GracePeriodSeconds: new(int64), PropagationPolicy: &replicaDeletion})
	if err != nil {
		return err
	}

	return nil
}

// DeleteApp deletes application
// 1. delete deployment
// 2. delete destination rule
// 3. delete virtual services
// 4. delete service
func DeleteApp(client kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery, appName string) error {
	// TODO wrap the error messages
	// TODO delete virtual services by hosts

	// 1. delete deployments
	replicaDeletion := metav1.DeletePropagationBackground
	client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).Delete(appName, &metav1.DeleteOptions{
		GracePeriodSeconds: new(int64), PropagationPolicy: &replicaDeletion,
	})

	// 2. delete destination rules
	err := istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Delete(appName, &metav1.DeleteOptions{})
	if err != nil {
		log.Println("fail to delete service: ", err)
	}

	// 3. delete virtual services
	err = istioClient.NetworkingV1alpha3().VirtualServices(namespace.ToRequestParam()).Delete(appName, &metav1.DeleteOptions{})
	if err != nil {
		log.Println("fail to delete virtual service: ", err)
	}

	// 4. delete service
	err = client.CoreV1().Services(namespace.ToRequestParam()).Delete(appName, &metav1.DeleteOptions{})
	if err != nil {
		log.Println("fail to delete service: ", err)
	}

	return nil
}

// removeFromDestinationRule removes the specified version from destination rule.
func removeFromDestinationRule(istioClient istio.Interface, rule *istioApi.DestinationRule, version string) error {
	subsets := []*istioApi.Subset{}
	for _, subset := range rule.Spec.Subsets {
		if subset.Name != version {
			subsets = append(subsets, subset)
		}
	}
	rule.Spec.Subsets = subsets

	_, err := istioClient.NetworkingV1alpha3().DestinationRules(rule.Namespace).Update(rule)
	return err
}

// removeFromVirtualService removes the specified subset from virtual service.
func removeFromVirtualService(istioClient istio.Interface, service *istioApi.VirtualService, subset string) error {
	for i := range service.Spec.Http {
		newRoute := []*istioApi.DestinationWeight{}
		for _, dest := range service.Spec.Http[i].Route {
			if dest.Destination.Subset != subset {
				newRoute = append(newRoute, dest)
			}
		}
		service.Spec.Http[i].Route = newRoute
	}

	_, err := istioClient.NetworkingV1alpha3().VirtualServices(service.Namespace).Update(service)
	return err
}

// TakeOverAllTraffic makes only one application's version is online.
// 1. make sure if this app's versioned destination rule exist
// 2. change virtual service to this version
func TakeOverAllTraffic(client kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery,
	appName string, version string, offlineType string) error {
	_, err := getDeploymentByLabels(client, namespace, map[string]string{
		"app":     appName,
		"version": version,
	})
	if err != nil {
		return err
	}

	var virtualServices []istioApi.VirtualService
	if virtualServices, err = virtualservice.GetVirtualServices(
		istioClient, []string{virtualservice.FQDN(appName, namespace.ToRequestParam())}, offlineType,
	); err != nil {
		return err
	}

	if len(virtualServices) == 0 { // no any virtualServices exists, create one
		vs := &istioApi.VirtualService{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.istio.io/v1alpha3",
				Kind:       "VirtualService",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      appName,
				Namespace: namespace.ToRequestParam(),
			},
			Spec: istioApi.VirtualServiceSpec{
				Hosts: []string{appName},
				Http: []*istioApi.HTTPRoute{
					{
						Route: []*istioApi.DestinationWeight{
							{
								Destination: &istioApi.Destination{
									Host:   appName,
									Subset: version,
								},
							},
						},
					},
				},
			},
		}
		_, err := istioClient.NetworkingV1alpha3().VirtualServices(namespace.ToRequestParam()).Create(vs)
		return err
	}

	for _, oldvsb := range virtualServices {
		oldvsb.Spec.Http = []*istioApi.HTTPRoute{
			{
				Route: []*istioApi.DestinationWeight{
					{
						Destination: &istioApi.Destination{
							Host:   appName,
							Subset: version,
						},
					},
				},
			},
		}

		_, err = istioClient.NetworkingV1alpha3().VirtualServices(namespace.ToRequestParam()).Update(&oldvsb)
		if err != nil {
			return err
		}
	}
	return nil
}
