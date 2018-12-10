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

	api2 "github.com/kubernetes/dashboard/src/app/backend/api"
	kdErrors "github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/destinationrule"
	"github.com/kubernetes/dashboard/src/app/backend/resource/namespace"
	"github.com/kubernetes/dashboard/src/app/backend/resource/service"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetApps queries application by namespace and data selector.
func GetApps(client kubernetes.Interface, istioClient istio.Interface, ns *common.NamespaceQuery, dataSelect *dataselect.DataSelectQuery) (*api.AppList, error) {
	dRules, err := destinationrule.GetDestinationRuleList(client, istioClient, ns, dataSelect)
	if err != nil {
		return nil, err
	}

	services, err := service.GetServiceList(client, ns, dataSelect)
	if err != nil {
		return nil, err
	}

	namespaces, err := namespace.GetNamespaceList(client, dataselect.NoDataSelect)
	if err != nil {
		return nil, err
	}

	// merge destinationRules & services
	result, err := getApps(services, dRules, nil, namespaces)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// getApps merges k8s services & istio destinationRules & virtualServices, return apps.
func getApps(services *service.ServiceList, dRules *destinationrule.List, vServices []istioApi.VirtualService, namespaces *namespace.NamespaceList) (*api.AppList, error) {
	var apps []*api.App

	if services == nil {
		return nil, errors.New("services not provided")
	}
	for _, svc := range services.Services {
		// use name+namespace to locate the app
		apps = append(apps, &api.App{
			ObjectMeta: svc.ObjectMeta,
			TypeMeta: api2.TypeMeta{
				Kind: api2.ResourceKindApp,
			},
			Destinations: []api.Destination{
				{
					Version:  "default",
					Selector: svc.Selector,
				},
			},
		})
	}

	if dRules != nil {
		for _, app := range apps {
			app.Destinations = getAppDestinations(app, dRules.DestinationRules)
		}
	}

	if vServices != nil {
		for _, app := range apps {
			app.VirtualServices = getAppVirtualServices(app, vServices)
		}
	}

	// add these applications' istio statuses
	if namespaces != nil {
		for _, app := range apps {
			ns := app.ObjectMeta.Namespace
			for _, n := range namespaces.Namespaces {
				if n.ObjectMeta.Name == ns {
					if n.ObjectMeta.Labels["istio-injection"] == "enabled" {
						app.Status.Istio = true
					}
					break
				}
			}
		}
	}

	list := &api.AppList{
		ListMeta: services.ListMeta,
		Apps:     apps,
	}
	return list, nil
}

func getAppVirtualServices(app *api.App, vServices []istioApi.VirtualService) []istioApi.VirtualService {
	var virtualServices = make([]istioApi.VirtualService, 0)

	if vServices == nil {
		return virtualServices
	}

	for _, vService := range vServices {
		var matched = false
		appAddr := virtualservice.FQDN(app.ObjectMeta.Name, app.ObjectMeta.Namespace)
		for _, vsHost := range vService.Spec.Hosts {
			svcAddr := virtualservice.FQDN(vsHost, vService.Namespace)
			if appAddr == svcAddr {
				virtualServices = append(virtualServices, vService)
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		for _, http := range vService.Spec.Http {
			for _, route := range http.Route {
				svcAddr := virtualservice.FQDN(route.Destination.Host, vService.Namespace)
				if appAddr == svcAddr {
					virtualServices = append(virtualServices, vService)
					break
				}
			}
		}
	}
	return virtualServices
}

func getAppDestinations(app *api.App, dRules []destinationrule.DestinationRuleDetail) []api.Destination {
	destinations := make([]api.Destination, 0)

	if dRules == nil {
		return destinations
	}

	for _, dRule := range dRules {
		svcAddr := virtualservice.FQDN(dRule.Host, dRule.ObjectMeta.Namespace)
		appAddr := virtualservice.FQDN(app.ObjectMeta.Name, app.ObjectMeta.Namespace)
		if appAddr == svcAddr {
			for _, sub := range dRule.Subsets {
				destinations = append(destinations, api.Destination{
					Version:  sub.Name,
					Selector: sub.Labels,
				})
			}
			break
		}
	}
	return destinations
}

// IsApp checks if the specified app, namespace is an app.
func IsApp(app string, namespace *common.NamespaceQuery) bool {
	return true
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
		TypeMeta: metaV1.TypeMeta{
			Kind:       api2.ResourceKindService,
			APIVersion: "v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
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
		TypeMeta: metaV1.TypeMeta{
			Kind:       api2.ResourceKindDeployment,
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metaV1.ObjectMeta{
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
			Selector: &metaV1.LabelSelector{
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
		ObjectMeta: metaV1.ObjectMeta{
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
		TypeMeta: metaV1.TypeMeta{
			Kind:       api2.ResourceKindDeployment,
			APIVersion: parent.APIVersion,
		},
		ObjectMeta: metaV1.ObjectMeta{
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
			Selector: &metaV1.LabelSelector{
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
	destinationRule, err := istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Get(appName, metaV1.GetOptions{})
	if err != nil {
		if kdErrors.IsNotFoundError(err) {
			// create a new destinationRule
			rule := &istioApi.DestinationRule{
				ObjectMeta: metaV1.ObjectMeta{
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

// fixDeployment fixes the deployment's labels & matchLabels according to the podTemplate.
// It checks whether the app & version labels exist for Istio running correctly.
func fixDeployment(client kubernetes.Interface, parent *v1beta1.Deployment) error {
	if parent.Spec.Template.Labels["app"] == "" || parent.Spec.Template.Labels["version"] == "" {
		return fmt.Errorf("parent deployment's pod template need to have app & version labels")
	}

	appName := parent.Spec.Template.Labels["app"]
	version := parent.Spec.Template.Labels["version"]

	if parent.Spec.Selector.MatchLabels["app"] == "" || parent.Spec.Selector.MatchLabels["version"] == "" ||
		parent.Labels["app"] == "" || parent.Labels["version"] == "" {
		parent.Labels["app"] = appName
		parent.Labels["version"] = version
		parent.Spec.Selector.MatchLabels["app"] = appName
		parent.Spec.Selector.MatchLabels["version"] = version
		_, err := client.ExtensionsV1beta1().Deployments(parent.Namespace).Update(parent)
		if err != nil {
			log.Println("fail to fix parent deployment", err)
			return err
		}
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
	replicaDeletion := metaV1.DeletePropagationBackground
	client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).Delete(appName, &metaV1.DeleteOptions{
		GracePeriodSeconds: new(int64), PropagationPolicy: &replicaDeletion,
	})

	// 2. delete destination rules
	err := istioClient.NetworkingV1alpha3().DestinationRules(namespace.ToRequestParam()).Delete(appName, &metaV1.DeleteOptions{})
	if err != nil {
		log.Println("fail to delete service: ", err)
	}

	// 3. delete virtual services
	err = istioClient.NetworkingV1alpha3().VirtualServices(namespace.ToRequestParam()).Delete(appName, &metaV1.DeleteOptions{})
	if err != nil {
		log.Println("fail to delete virtual service: ", err)
	}

	// 4. delete service
	err = client.CoreV1().Services(namespace.ToRequestParam()).Delete(appName, &metaV1.DeleteOptions{})
	if err != nil {
		log.Println("fail to delete service: ", err)
	}

	return nil
}

// addToDestinationRule adds the specified version from destination rule.
func addToDestinationRule(client istio.Interface, rule *istioApi.DestinationRule, version string, namespace string) error {
	// TODO the same reason as destination rule
	subsets := []*istioApi.Subset{}
	for _, subset := range rule.Spec.Subsets {
		if subset.Name != version {
			subsets = append(subsets, subset)
		} else {
			return nil
		}
	}

	subsets = append(subsets, &istioApi.Subset{
		Name: version,
		Labels: map[string]string{
			"version": version,
		},
	})
	rule.Spec.Subsets = subsets
	_, err := client.NetworkingV1alpha3().DestinationRules(namespace).Update(rule)
	return err
}
