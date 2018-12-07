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
	"fmt"
	"log"
	"time"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func OfflineAppVersion(client kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery,
	appName string, version string, offlineType string) error {
	var (
		virtualServices []istioApi.VirtualService
		err             error
	)

	virtualServices, err = virtualservice.GetVirtualServices(istioClient, []string{appName}, offlineType)

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
	for i, _ := range service.Spec.Http {
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
