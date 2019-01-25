package v1alpha3

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceEntry describes istio's flow control.
type ServiceEntry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceEntrySpec `json:"spec"`
}

type ServiceEntrySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// REQUIRED. The hosts associated with the ServiceEntry. Could be a DNS
	// name with wildcard prefix (external services only). DNS names in hosts
	// will be ignored if the application accesses the service over non-HTTP
	// protocols such as mongo/opaque TCP/even HTTPS. In such scenarios, the
	// IP addresses specified in the Addresses field or the port will be used
	// to uniquely identify the destination.
	Hosts []string `protobuf:"bytes,1,rep,name=hosts" json:"hosts,omitempty"`
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
	Addresses []string `protobuf:"bytes,2,rep,name=addresses" json:"addresses,omitempty"`
	// REQUIRED. The ports associated with the external service. If the
	// Endpoints are unix domain socket addresses, there must be exactly one
	// port.
	Ports []*Port `protobuf:"bytes,3,rep,name=ports" json:"ports,omitempty"`
	// Specify whether the service should be considered external to the mesh
	// or part of the mesh.
	Location int32 `protobuf:"varint,4,opt,name=location,proto3,enum=istio.networking.v1alpha3.ServiceEntry_Location" json:"location,omitempty"`
	// REQUIRED: Service discovery mode for the hosts. Care must be taken
	// when setting the resolution mode to NONE for a TCP port without
	// accompanying IP addresses. In such cases, traffic to any IP on
	// said port will be allowed (i.e. 0.0.0.0:<port>).
	Resolution int32 `protobuf:"varint,5,opt,name=resolution,proto3,enum=istio.networking.v1alpha3.ServiceEntry_Resolution" json:"resolution,omitempty"`
	// One or more endpoints associated with the service.
	Endpoints []*ServiceEntry_Endpoint `protobuf:"bytes,6,rep,name=endpoints" json:"endpoints,omitempty"`
}

type ServiceEntry_Endpoint struct {
	// REQUIRED: Address associated with the network endpoint without the
	// port.  Domain names can be used if and only if the resolution is set
	// to DNS, and must be fully-qualified without wildcards. Use the form
	// unix:///absolute/path/to/socket for unix domain socket endpoints.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// Set of ports associated with the endpoint. The ports must be
	// associated with a port name that was declared as part of the
	// service. Do not use for unix:// addresses.
	Ports map[string]uint32 `protobuf:"bytes,2,rep,name=ports" json:"ports,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// One or more labels associated with the endpoint.
	Labels map[string]string `protobuf:"bytes,3,rep,name=labels" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceEntryList is a list of ServiceEntry resources
type ServiceEntryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceEntry `json:"items"`
}
