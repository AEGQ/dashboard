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

package ingress

import (
	"log"

	commonApi "github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetIngress returns the specified istio ingress.
func GetIngress(k8sClient kubernetes.Interface, istioClient istio.Interface,
	namespace, name string) (*api.Ingress, error) {

	log.Printf("Getting details of %s ingress in %s namespace", name, namespace)

	gateway, err := istioClient.NetworkingV1alpha3().Gateways(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	selector, err := metaV1.LabelSelectorAsSelector(&metaV1.LabelSelector{MatchLabels: gateway.Spec.Selector})
	if err != nil {
		return nil, err
	}
	options := metaV1.ListOptions{LabelSelector: selector.String()}
	channels := &common.ResourceChannels{
		ServiceList: common.GetServiceListChannelWithOptions(k8sClient, common.NewSameNamespaceQuery(namespace),
			options, 1),
		VirtualServiceList: common.GetVirtualServiceListChannel(istioClient, common.NewSameNamespaceQuery(namespace), 1),
	}

	rawServices := <-channels.ServiceList.List
	err = <-channels.ServiceList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	rawVs := <-channels.VirtualServiceList.List
	err = <-channels.VirtualServiceList.Error
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	return ToIngress(gateway, rawServices.Items, rawVs.Items, nonCriticalErrors), nil
}

func ToIngress(gateway *v1alpha3.Gateway, services []v1.Service, virtualServices []v1alpha3.VirtualService, nonCriticalErrors []error) *api.Ingress {
	var hosts []common.Endpoint
	var hostMapping = make(map[string][]common.ServicePort, 0)
	for _, server := range gateway.Spec.Servers {
		for _, h := range server.Hosts {
			if _, exist := hostMapping[h]; exist {
				hostMapping[h] = append(hostMapping[h], common.ServicePort{
					Port: int32(server.Port.Number),
				})
			} else {
				hostMapping[h] = []common.ServicePort{
					{
						Port: int32(server.Port.Number),
					},
				}
			}
		}
	}
	for h, p := range hostMapping {
		hosts = append(hosts, common.Endpoint{
			Host:  h,
			Ports: p,
		})
	}

	var externalEndpoints = make([]common.Endpoint, 0)
	svcs := filterServiceByLabels(gateway.Spec.Selector, services)
	for _, svc := range svcs {
		externalEndpoints = append(externalEndpoints, common.GetExternalEndpoints(svc)...)
	}

	var filteredVs []virtualservice.VirtualService
	if virtualServices != nil {
		for _, vs := range virtualServices {
			if contains(vs.Spec.Gateways, gateway.Name) {
				filteredVs = append(filteredVs, virtualservice.ToVirtualServiceDetail(&vs, nil, nil, nil))
			}
		}
	}

	return &api.Ingress{
		ObjectMeta:        commonApi.NewObjectMeta(gateway.ObjectMeta),
		TypeMeta:          commonApi.NewTypeMeta(commonApi.ResourceKindIngress),
		Hosts:             hosts,
		ExternalEndpoints: externalEndpoints,
		VirtualServices:   filteredVs,
		Errors:            nonCriticalErrors,
	}
}

func contains(collection []string, element string) bool {
	for _, e := range collection {
		if e == element {
			return true
		}
	}
	return false
}
