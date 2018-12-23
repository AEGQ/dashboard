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
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	v1 "k8s.io/api/core/v1"
)

type GatewayCell istioApi.Gateway

func ToGateway(g *v1alpha3.Gateway) *Gateway {
	endpoints := make([]*common.Endpoint, 0)
	for _, s := range g.Spec.Servers {
		for _, h := range s.Hosts {
			endpoints = append(endpoints, &common.Endpoint{
				Host: h,
				Ports: []common.ServicePort{
					{
						Port:     int32(s.Port.Number),
						Protocol: v1.Protocol(s.Port.Protocol),
					},
				},
			})
		}
	}
	return &Gateway{
		ObjectMeta: api.NewObjectMeta(g.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindGateway),
		Servers:    g.Spec.Servers,
		Selector:   g.Spec.Selector,
		Endpoints:  endpoints,
	}
}

func (self GatewayCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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
