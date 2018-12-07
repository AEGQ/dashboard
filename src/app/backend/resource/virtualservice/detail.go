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

package virtualservice

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/destinationrule"
	"github.com/kubernetes/dashboard/src/app/backend/resource/service"
	resourceService "github.com/kubernetes/dashboard/src/app/backend/resource/service"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// VirtualService is a representation of a kubernetes VirtualService object.
type VirtualService struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	// Hosts include the hosts virtual service apply to.
	Hosts []string `json:"hosts"`
	// The names of gateways and sidecars that should apply these routes. A
	// single VirtualService is used for sidecars inside the mesh as well as
	// for one or more gateways.
	Gateways []string `json:"gateways,omitempty"`
	// Http represents the http route rules.
	Http []*istioApi.HTTPRoute `json:"http,omitempty"`
	// The tls match.
	Tls []*istioApi.TLSRoute `json:"tls,omitempty"`
	// Tcp route rules.
	Tcp []*istioApi.TCPRoute `json:"tcp,omitempty"`
	// The destination rules belong to the virtual service's hosts
	DestinationRuleList *destinationrule.List `json:"destinationRuleList,omitempty"`
	// Service is the related k8s service with the virtual service
	ServiceList *service.ServiceList `json:"serviceList,omitempty"`
	// ServiceEntry is istio extended service for outbound service or out-of-k8s service
	ServiceEntryList string `json:"serviceEntryList,omitempty"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

func GetVirtualService(k8sClient kubernetes.Interface, istioClient istio.Interface, name, namespace string) (*VirtualService, error) {
	log.Printf("Getting details of %s virtual service", name)

	virtualService, err := istioClient.NetworkingV1alpha3().VirtualServices(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var nonCriticalErrors = make([]error, 0)
	var criticalError error
	hosts := virtualService.Spec.Hosts

	// fetch related destination rule
	destinationList, err := destinationrule.GetDestinationRuleListByHostname(k8sClient, istioClient, common.NewNamespaceQuery([]string{namespace}), hosts)
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	// TODO optimize related service query
	//dataSelect := parseDataSelectPathParameter(request)
	//namespace := parseNamespacePathParameter(request)
	//hosts := vs.Hosts
	services, err := resourceService.GetServiceList(k8sClient, common.NewNamespaceQuery([]string{namespace}), dataselect.NoDataSelect)
	var newSvcs []service.Service
	for _, svc := range services.Services {
		for _, h := range hosts {
			if svc.ObjectMeta.Name == h {
				newSvcs = append(newSvcs, svc)
				break
			}
		}
	}
	services.ListMeta.TotalItems = len(newSvcs)
	services.Services = newSvcs
	nonCriticalErrors, criticalError = errors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	// TODO fetch related service entries

	detail := ToVirtualServiceDetail(virtualService, destinationList, services, nonCriticalErrors)
	return &detail, nil
}
