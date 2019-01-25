package v1alpha3

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Gateway describes istio's flow control.
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GatewaySpec `json:"spec"`
}

type GatewaySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Servers  []*Server         `protobuf:"bytes,1,rep,name=servers" json:"servers,omitempty"`
	Selector map[string]string `protobuf:"bytes,2,rep,name=selector" json:"selector,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

type Server struct {
	// REQUIRED: The Port on which the proxy should listen for incoming
	// connections
	Port *Port `protobuf:"bytes,1,opt,name=port" json:"port,omitempty"`
	// REQUIRED. A list of hosts exposed by this gateway. At least one
	// host is required. While typically applicable to
	// HTTP services, it can also be used for TCP services using TLS with
	// SNI. May contain a wildcard prefix for the bottom-level component of
	// a domain name. For example `*.foo.com` matches `bar.foo.com`
	// and `*.com` matches `bar.foo.com`, `example.com`, and so on.
	//
	// **Note**: A `VirtualService` that is bound to a gateway must have one
	// or more hosts that match the hosts specified in a server. The match
	// could be an exact match or a suffix match with the server's hosts. For
	// example, if the server's hosts specifies "*.example.com",
	// VirtualServices with hosts dev.example.com, prod.example.com will
	// match. However, VirtualServices with hosts example.com or
	// newexample.com will not match.
	Hosts []string `protobuf:"bytes,2,rep,name=hosts" json:"hosts,omitempty"`
	// Set of TLS related options that govern the server's behavior. Use
	// these options to control if all http requests should be redirected to
	// https, and the TLS modes to use.
	Tls *Server_TLSOptions `protobuf:"bytes,3,opt,name=tls" json:"tls,omitempty"`
}

type Server_TLSOptions struct {
	// If set to true, the load balancer will send a 301 redirect for all
	// http connections, asking the clients to use HTTPS.
	HttpsRedirect bool `protobuf:"varint,1,opt,name=https_redirect,json=httpsRedirect,proto3" json:"https_redirect,omitempty"`
	// Optional: Indicates whether connections to this port should be
	// secured using TLS. The value of this field determines how TLS is
	// enforced.
	Mode string `protobuf:"varint,2,opt,name=mode,proto3,enum=istio.networking.v1alpha3.Server_TLSOptions_TLSmode" json:"mode,omitempty"`
	// REQUIRED if mode is `SIMPLE` or `MUTUAL`. The path to the file
	// holding the server-side TLS certificate to use.
	ServerCertificate string `protobuf:"bytes,3,opt,name=server_certificate,json=serverCertificate,proto3" json:"server_certificate,omitempty"`
	// REQUIRED if mode is `SIMPLE` or `MUTUAL`. The path to the file
	// holding the server's private key.
	PrivateKey string `protobuf:"bytes,4,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	// REQUIRED if mode is `MUTUAL`. The path to a file containing
	// certificate authority certificates to use in verifying a presented
	// client side certificate.
	CaCertificates string `protobuf:"bytes,5,opt,name=ca_certificates,json=caCertificates,proto3" json:"ca_certificates,omitempty"`
	// A list of alternate names to verify the subject identity in the
	// certificate presented by the client.
	SubjectAltNames []string `protobuf:"bytes,6,rep,name=subject_alt_names,json=subjectAltNames" json:"subject_alt_names,omitempty"`
}

type Port struct {
	// REQUIRED: A valid non-negative integer port number.
	Number uint32 `protobuf:"varint,1,opt,name=number,proto3" json:"number,omitempty"`
	// REQUIRED: The protocol exposed on the port.
	// MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS.
	// TLS is used to indicate secure connections to non HTTP services.
	Protocol string `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`
	// Label assigned to the port.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayList is a list of Gateway resources
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Gateway `json:"items"`
}
