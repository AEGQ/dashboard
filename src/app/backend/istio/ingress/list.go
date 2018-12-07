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
	commonApi "github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetIngresses merges istio `Gateway`, `VirtualService`, `Service`.
// 1. Query istio `Gateway`s.
// 2. Map Gateway with `Deployment`'s Services.
func GetIngresses(k8sClient kubernetes.Interface, istioClient istio.Interface,
	nsQuery *common.NamespaceQuery, dataSelect *dataselect.DataSelectQuery) (*api.IngressList, error) {
	channels := common.ResourceChannels{
		GatewayList: common.GetGatewayListChannel(istioClient, nsQuery, 1),
		ServiceList: common.GetServiceListChannel(k8sClient, nsQuery, 1),
	}

	return GetIngressesFromChannel(&channels)
}

func GetIngressesFromChannel(channels *common.ResourceChannels) (*api.IngressList, error) {
	gateways := <-channels.GatewayList.List
	err := <-channels.GatewayList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	services := <-channels.ServiceList.List
	err = <-channels.ServiceList.Error
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	return ToIngressList(gateways.Items, services.Items, nonCriticalErrors), nil
}

func ToIngressList(gateways []v1alpha3.Gateway, services []v1.Service, nonCriticalErrors []error) *api.IngressList {
	ingressList := &api.IngressList{
		Items:    make([]api.Ingress, 0),
		ListMeta: commonApi.ListMeta{TotalItems: len(gateways)},
		Errors:   nonCriticalErrors,
	}

	for _, gateway := range gateways {
		ingressList.Items = append(ingressList.Items, *ToIngress(&gateway, services, nil))
	}

	return ingressList
}

func filterServiceByLabels(labels map[string]string, services []v1.Service) []*v1.Service {
	var results []*v1.Service
	for _, svc := range services {
		if subsetOf(labels, svc.Labels) {
			results = append(results, &svc)
		}
	}
	return results
}

func subsetOf(given map[string]string, target map[string]string) bool {
	for k, v := range given {
		if target[k] != v {
			return false
		}
	}
	return true
}

func CreateIngress(k8sClient kubernetes.Interface, istioClient istio.Interface,
	namespace *common.NamespaceQuery, name string) error {
	// TODO
	return nil
}

func DeleteIngress(k8sClient kubernetes.Interface, istioClient istio.Interface,
	namespace, name string) error {
	// TODO
	return nil
}
