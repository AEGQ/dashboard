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
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Gateway struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
	// REQUIRED: A list of server specifications.
	Servers []*istioApi.Server `json:"servers,omitempty"`
	// REQUIRED: One or more labels that indicate a specific set of pods/VMs
	// on which this gateway configuration should be applied.
	// The scope of label search is platform dependent.
	// On Kubernetes, for example, the scope includes pods running in
	// all reachable namespaces.
	Selector  map[string]string  `json:"selector,omitempty"`
	Endpoints []*common.Endpoint `json:"endpoints"`
}

type Server struct {
	Hosts []string          `json:"hosts,omitempty"`
	Port  *istioApi.Port    `json:"port,omitempty"`
	Tls   Server_TLSOptions `json:"tls,omitempty"`
}

type Server_TLSOptions struct {
	Mode              string `json:"mode,omitempty"`
	PrivateKey        string `json:"privateKey,omitempty"`
	ServerCertificate string `json:"serverCertificate,omitempty"`
}

func GetGateway(k8sClient kubernetes.Interface, istioClient istio.Interface, name, namespace string) (*Gateway, error) {
	gateway, err := istioClient.NetworkingV1alpha3().Gateways(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ToGateway(gateway), nil
}
