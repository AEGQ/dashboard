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

	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
)

func getAppVirtualServices(app *api.App, vServices []istioApi.VirtualService) []virtualservice.VirtualService {
	var virtualServices = make([]virtualservice.VirtualService, 0)

	if vServices == nil {
		return virtualServices
	}

	for _, vService := range vServices {
		var matched = false
		appAddr := virtualservice.FQDN(app.ObjectMeta.Name, app.ObjectMeta.Namespace)
		for _, vsHost := range vService.Spec.Hosts {
			svcAddr := virtualservice.FQDN(vsHost, vService.Namespace)
			if appAddr == svcAddr {
				virtualServices = append(virtualServices, virtualservice.ToVirtualServiceDetail(&vService, nil, nil, nil))
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
					virtualServices = append(virtualServices, virtualservice.ToVirtualServiceDetail(&vService, nil, nil, nil))
					break
				}
			}
		}
	}
	return virtualServices
}

func getAppDestinations(app *api.App, dRules []istioApi.DestinationRule) []api.Destination {
	destinations := make([]api.Destination, 0)

	if dRules == nil {
		return destinations
	}

	for _, dRule := range dRules {
		svcAddr := virtualservice.FQDN(dRule.Spec.Host, dRule.ObjectMeta.Namespace)
		appAddr := virtualservice.FQDN(app.ObjectMeta.Name, app.ObjectMeta.Namespace)
		if appAddr == svcAddr {
			for _, sub := range dRule.Spec.Subsets {
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
