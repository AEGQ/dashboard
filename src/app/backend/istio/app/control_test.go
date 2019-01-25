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

package app

import (
	"testing"

	"github.com/magiconair/properties/assert"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTakeOver(t *testing.T) {
	vs := v1alpha3.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind: api.ResourceKindVirtualService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-test-com",
			Namespace: "wall",
		},
		Spec: v1alpha3.VirtualServiceSpec{
			Hosts:    []string{"api.test.com"},
			Gateways: []string{"api-test-com"},
			Http: []*v1alpha3.HTTPRoute{
				{
					Match: []*v1alpha3.HTTPMatchRequest{},
					Route: []*v1alpha3.DestinationWeight{
						{
							Destination: &v1alpha3.Destination{
								Host:   "test",
								Subset: "v3",
								Port: &v1alpha3.PortSelector{
									Number: 8080,
								},
							},
							Weight: 100,
						},
						{
							Destination: &v1alpha3.Destination{
								Host:   "test.wall.svc.cluster.local",
								Subset: "v2",
							},
							Weight: 100,
						},
						{
							Destination: &v1alpha3.Destination{
								Host:   "something_else",
								Subset: "v3",
								Port: &v1alpha3.PortSelector{
									Number: 8080,
								},
							},
							Weight: 100,
						},
					},
				},
			},
		},
	}

	overrideSubset(&vs, "test", "wall", "v1")

	assert.Equal(t, vs.Spec.Http[0].Route[0].Destination.Host, "test")
	assert.Equal(t, vs.Spec.Http[0].Route[0].Destination.Subset, "v1")

	assert.Equal(t, vs.Spec.Http[0].Route[1].Destination.Host, "test.wall.svc.cluster.local")
	assert.Equal(t, vs.Spec.Http[0].Route[1].Destination.Subset, "v1")

	assert.Equal(t, vs.Spec.Http[0].Route[2].Destination.Host, "something_else")
	assert.Equal(t, vs.Spec.Http[0].Route[2].Destination.Subset, "v3")
}
