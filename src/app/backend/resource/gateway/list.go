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

package gateway

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type List struct {
	ListMeta api.ListMeta `json:"listMeta"`
	Items    []Gateway    `json:"gateways"`
	Errors   []error      `json:"errors"`
}

func GetGatewayList(k8sClient kubernetes.Interface, istioClient istio.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	channels := &common.ResourceChannels{
		GatewayList: common.GetGatewayListChannel(istioClient, nsQuery, 1),
	}

	return GetGatewayListFromChannels(channels, dsQuery)
}

func GetGatewayListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*List, error) {

	gateways := <-channels.GatewayList.List
	err := <-channels.GatewayList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	GatewayList := toGatewayList(gateways.Items, nonCriticalErrors, dsQuery)
	return GatewayList, nil
}

func toGatewayList(gateways []v1alpha3.Gateway, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *List {
	gatewayList := &List{
		Items:    make([]Gateway, 0),
		ListMeta: api.ListMeta{TotalItems: len(gateways)},
		Errors:   nonCriticalErrors,
	}

	gatewayCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(gateways), dsQuery)
	gateways = fromCells(gatewayCells)

	for _, gateway := range gateways {
		gatewayList.Items = append(gatewayList.Items, *ToGateway(&gateway))
	}

	gatewayList.ListMeta.TotalItems = filteredTotal

	return gatewayList
}

func fromCells(cells []dataselect.DataCell) []istioApi.Gateway {
	std := make([]istioApi.Gateway, len(cells))
	for i := range std {
		std[i] = istioApi.Gateway(cells[i].(GatewayCell))
	}
	return std
}

func toCells(std []istioApi.Gateway) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = GatewayCell(std[i])
	}
	return cells
}
