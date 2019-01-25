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
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DestinationRuleDetail is a representation of a kubernetes DestinationRuleDetail object.
type DestinationRuleDetail struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	Host          string                  `json:"host,omitempty"`
	TrafficPolicy *istioApi.TrafficPolicy `json:"traffic_policy,omitempty"`
	Subsets       []*istioApi.Subset      `json:"subsets,omitempty"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetDestinationRule returns destination rule object.
func GetDestinationRule(k8sClient kubernetes.Interface, istioClient istio.Interface, name string, namespace string) (*DestinationRuleDetail, error) {
	destinationRule, err := istioClient.NetworkingV1alpha3().DestinationRules(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ToDestinationRuleDetail(destinationRule, nil), nil
}
