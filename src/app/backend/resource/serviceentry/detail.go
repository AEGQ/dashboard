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
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceEntry struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
	// REQUIRED. The hosts associated with the ServiceEntry. Could be a DNS
	// name with wildcard prefix (external services only). DNS names in hosts
	// will be ignored if the application accesses the service over non-HTTP
	// protocols such as mongo/opaque TCP/even HTTPS. In such scenarios, the
	// IP addresses specified in the Addresses field or the port will be used
	// to uniquely identify the destination.
	Hosts []string `json:"hosts,omitempty"`
	// The virtual IP addresses associated with the service. Could be CIDR
	// prefix.  For HTTP services, the addresses field will be ignored and
	// the destination will be identified based on the HTTP Host/Authority
	// header. For non-HTTP protocols such as mongo/opaque TCP/even HTTPS,
	// the hosts will be ignored. If one or more IP addresses are specified,
	// the incoming traffic will be identified as belonging to this service
	// if the destination IP matches the IP/CIDRs specified in the addresses
	// field. If the Addresses field is empty, traffic will be identified
	// solely based on the destination port. In such scenarios, the port on
	// which the service is being accessed must not be shared by any other
	// service in the mesh. In other words, the sidecar will behave as a
	// simple TCP proxy, forwarding incoming traffic on a specified port to
	// the specified destination endpoint IP/host. Unix domain socket
	// addresses are not supported in this field.
	Addresses []string `json:"addresses,omitempty"`
	// REQUIRED. The ports associated with the external service. If the
	// Endpoints are unix domain socket addresses, there must be exactly one
	// port.
	Ports []*v1alpha3.Port `json:"ports,omitempty"`
	// TODO more fields
	// One or more endpoints associated with the service.
	Endpoints []*v1alpha3.ServiceEntry_Endpoint `json:"endpoints,omitempty"`
}

func GetServiceEntry(k8sClient kubernetes.Interface, istioClient istio.Interface, name, namespace string) (*ServiceEntry, error) {
	serviceEntry, err := istioClient.NetworkingV1alpha3().ServiceEntries(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ToServiceEntry(serviceEntry), nil
}
