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
	"fmt"
	"log"
	"strings"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	OnlyHost    string = "host"
	OnlyGateway        = "gateway"
	All                = "all"
)

type List struct {
	ListMeta        api.ListMeta     `json:"listMeta"`
	VirtualServices []VirtualService `json:"virtualServices"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetVirtualServiceList fetches virtual services under specified namespace.
func GetVirtualServiceList(k8sClient kubernetes.Interface, istioClient istio.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	log.Printf("Getting details of virtual services under %s", nsQuery.ToRequestParam())
	channels := &common.ResourceChannels{
		VirtualServiceList: common.GetVirtualServiceListChannel(istioClient, nsQuery, 1),
	}
	return GetVirtualServiceListFromChannel(channels, dsQuery)
}

func GetVirtualServiceListFromChannel(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	virtualServices := <-channels.VirtualServiceList.List
	err := <-channels.VirtualServiceList.Error

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return ToVirtualServiceList(virtualServices.Items, dsQuery, nonCriticalErrors), nil
}

func ToVirtualServiceList(virtualServices []v1alpha3.VirtualService, dsQuery *dataselect.DataSelectQuery, nonCriticalErrors []error) *List {
	virtualServiceList := &List{
		VirtualServices: make([]VirtualService, 0),
		ListMeta:        api.ListMeta{TotalItems: len(virtualServices)},
		Errors:          nonCriticalErrors,
	}

	virtualServiceCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(virtualServices), dsQuery)
	virtualServices = fromCells(virtualServiceCells)

	for _, virtualService := range virtualServices {
		virtualServiceList.VirtualServices = append(virtualServiceList.VirtualServices, ToVirtualServiceDetail(&virtualService, nil, nil, nonCriticalErrors))
	}

	virtualServiceList.ListMeta.TotalItems = filteredTotal
	return virtualServiceList
}

func fromCells(cells []dataselect.DataCell) []istioApi.VirtualService {
	std := make([]istioApi.VirtualService, len(cells))
	for i := range std {
		std[i] = istioApi.VirtualService(cells[i].(VirtualServiceCell))
	}
	return std
}

func toCells(std []istioApi.VirtualService) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = VirtualServiceCell(std[i])
	}
	return cells
}

func intersected(col1 []string, col2 []string) bool {
	for _, c1 := range col1 {
		for _, c2 := range col2 {
			if c1 == c2 {
				return true
			}
		}
	}
	return false
}

// GetVirtualServices fetches virtualServices under specified namespace by filtering hosts.
func GetVirtualServices(istioClient istio.Interface, hosts []string, targetType string) ([]istioApi.VirtualService, error) {
	var vServices = make([]istioApi.VirtualService, 0)

	virtualServices, err := istioClient.NetworkingV1alpha3().VirtualServices("").
		List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range virtualServices.Items {
		if targetType == OnlyHost || targetType == All {
			if intersected(toFQDNs(item.Spec.Hosts, item.Namespace), hosts) {
				vServices = append(vServices, item)
				continue
			}
		}
		if targetType == OnlyGateway || targetType == All {
			if intersected(getDestinations(item), hosts) {
				vServices = append(vServices, item)
				continue
			}
		}
	}

	return vServices, nil
}

func toFQDNs(hosts []string, namespace string) []string {
	newHosts := make([]string, 0)
	for _, h := range hosts {
		newHosts = append(newHosts, FQDN(h, namespace))
	}
	return newHosts
}

func getDestinations(v istioApi.VirtualService) []string {
	destinationHosts := []string{}
	for _, r := range v.Spec.Http {
		for _, d := range r.Route {
			destinationHosts = append(destinationHosts, FQDN(d.Destination.Host, v.Namespace))
		}
	}
	return destinationHosts
}

// FQDN interprets the service host to full qualified domain name.
func FQDN(svc string, namespace string) string {
	if strings.HasSuffix(svc, "svc.cluster.local") {
		return svc
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local", svc, namespace)
}
