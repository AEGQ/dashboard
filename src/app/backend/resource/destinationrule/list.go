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

package destinationrule

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

	DestinationRules []DestinationRuleDetail `json:"destinationRules"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

func GetDestinationRuleList(k8sClient kubernetes.Interface, istioClient istio.Interface, nsQuery *common.NamespaceQuery, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	channels := &common.ResourceChannels{
		DestinationRuleList: common.GetDestinationRuleListChannel(istioClient, nsQuery, 1),
	}

	return GetDestinationRuleListFromChannels(channels, dsQuery)
}

func GetDestinationRuleListFromChannels(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*List, error) {
	destinationRules := <-channels.DestinationRuleList.List
	err := <-channels.DestinationRuleList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return ToDestinationRuleList(destinationRules.Items, nonCriticalErrors, dsQuery), nil
}

func ToDestinationRuleList(destinationRules []v1alpha3.DestinationRule, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *List {
	destinationRuleList := &List{
		DestinationRules: make([]DestinationRuleDetail, 0),
		ListMeta:         api.ListMeta{TotalItems: len(destinationRules)},
		Errors:           nonCriticalErrors,
	}

	destinationRuleCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(destinationRules), dsQuery)
	destinationRules = fromCells(destinationRuleCells)

	for _, destinationRule := range destinationRules {
		destinationRuleList.DestinationRules = append(destinationRuleList.DestinationRules, *ToDestinationRuleDetail(&destinationRule, nonCriticalErrors))
	}

	destinationRuleList.ListMeta.TotalItems = filteredTotal

	return destinationRuleList
}

func fromCells(cells []dataselect.DataCell) []istioApi.DestinationRule {
	std := make([]istioApi.DestinationRule, len(cells))
	for i := range std {
		std[i] = istioApi.DestinationRule(cells[i].(DestinationRuleCell))
	}
	return std
}

func toCells(std []istioApi.DestinationRule) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = DestinationRuleCell(std[i])
	}
	return cells
}

func GetDestinationRuleListByHostname(k8sClient kubernetes.Interface, istioClient istio.Interface, namespace *common.NamespaceQuery, hostnames []string) (*List, error) {
	list, err := GetDestinationRuleList(k8sClient, istioClient, namespace, dataselect.NoDataSelect)
	if err != nil {
		return nil, err
	}

	var rules []DestinationRuleDetail
	for _, item := range list.DestinationRules {
		for _, hostname := range hostnames {
			if hostname == item.Host {
				rules = append(rules, item)
			}
		}
	}

	return &List{
		ListMeta: api.ListMeta{
			TotalItems: len(rules),
		},
		DestinationRules: rules,
	}, nil
}
