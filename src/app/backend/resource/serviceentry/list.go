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

package serviceentry

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type List struct {
	ListMeta       api.ListMeta   `json:"listMeta"`
	ServiceEntries []ServiceEntry `json:"serviceEntries"`
	Errors         []error        `json:"errors"`
}

func GetServiceEntryList(k8sClient kubernetes.Interface, istioClient istio.Interface,
	namespace *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	channels := &common.ResourceChannels{
		ServiceEntryList: common.GetServiceEntryListChannel(istioClient, namespace, 1),
	}

	return GetServiceEntryListFromChannel(channels, dsQuery)
}

func GetServiceEntryListFromChannel(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	serviceEntries := <-channels.ServiceEntryList.List
	err := <-channels.ServiceEntryList.Error

	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return ToServiceEntryList(serviceEntries.Items, dsQuery, nonCriticalErrors)
}

func ToServiceEntryList(serviceEntries []v1alpha3.ServiceEntry, dsQuery *dataselect.DataSelectQuery,
	nonCriticalErrors []error) (*List, error) {
	serviceEntryList := &List{
		ListMeta:       api.ListMeta{TotalItems: 0},
		ServiceEntries: make([]ServiceEntry, 0),
		Errors:         nonCriticalErrors,
	}

	serviceEntryCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(serviceEntries), dsQuery)
	serviceEntries = fromCells(serviceEntryCells)

	for _, serviceEntry := range serviceEntries {
		serviceEntryList.ServiceEntries = append(serviceEntryList.ServiceEntries, *ToServiceEntry(&serviceEntry))
	}

	serviceEntryList.ListMeta.TotalItems = filteredTotal

	return serviceEntryList, nil
}

func fromCells(cells []dataselect.DataCell) []v1alpha3.ServiceEntry {
	std := make([]v1alpha3.ServiceEntry, len(cells))
	for i := range std {
		std[i] = v1alpha3.ServiceEntry(cells[i].(ServiceEntryCell))
	}
	return std
}

func toCells(std []v1alpha3.ServiceEntry) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = ServiceEntryCell(std[i])
	}
	return cells
}
