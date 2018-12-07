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
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
)

type DestinationRuleCell istioApi.DestinationRule

// ToDestinationRuleDetail returns api service object based on kubernetes destination rule object
func ToDestinationRuleDetail(d *istioApi.DestinationRule, nonCriticalErrors []error) *DestinationRuleDetail {
	return &DestinationRuleDetail{
		ObjectMeta:    api.NewObjectMeta(d.ObjectMeta),
		TypeMeta:      api.NewTypeMeta(api.ResourceKindDestinationRule),
		Host:          d.Spec.Host,
		TrafficPolicy: d.Spec.TrafficPolicy,
		Subsets:       d.Spec.Subsets,
		Errors:        nonCriticalErrors,
	}
}

func (self DestinationRuleCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}
