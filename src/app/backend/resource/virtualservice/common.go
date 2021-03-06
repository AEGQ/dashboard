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
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/destinationrule"
	"github.com/kubernetes/dashboard/src/app/backend/resource/service"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
)

type VirtualServiceCell istioApi.VirtualService

func ToVirtualServiceDetail(vs *istioApi.VirtualService, destinationRules *destinationrule.List, services *service.ServiceList, nonCriticalErrors []error) VirtualService {
	return VirtualService{
		ObjectMeta:          api.NewObjectMeta(vs.ObjectMeta),
		TypeMeta:            api.NewTypeMeta(api.ResourceKindVirtualService),
		Hosts:               vs.Spec.Hosts,
		Gateways:            vs.Spec.Gateways,
		Http:                vs.Spec.Http,
		Tls:                 vs.Spec.Tls,
		Tcp:                 vs.Spec.Tcp,
		DestinationRuleList: destinationRules,
		ServiceList:         services,
		Errors:              nonCriticalErrors,
	}
}

func (self VirtualServiceCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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
