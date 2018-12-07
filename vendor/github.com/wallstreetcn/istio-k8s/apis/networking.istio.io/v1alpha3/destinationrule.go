package v1alpha3

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DestinationRule describes istio's flow control.
type DestinationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DestinationRuleSpec `json:"spec"`
}

type DestinationRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Host          string         `json:"host,omitempty"`
	TrafficPolicy *TrafficPolicy `protobuf:"bytes,2,opt,name=traffic_policy,json=trafficPolicy" json:"traffic_policy,omitempty"`
	// One or more named sets that represent individual versions of a
	// service. Traffic policies can be overridden at subset level.
	Subsets []*Subset `protobuf:"bytes,3,rep,name=subsets" json:"subsets,omitempty"`
}

type Subset struct {
	// REQUIRED. Name of the subset. The service name and the subset name can
	// be used for traffic splitting in a route rule.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// REQUIRED. Labels apply a filter over the endpoints of a service in the
	// service registry. See route rules for examples of usage.
	Labels map[string]string `protobuf:"bytes,2,rep,name=labels" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Traffic policies that apply to this subset. Subsets inherit the
	// traffic policies specified at the DestinationRule level. Settings
	// specified at the subset level will override the corresponding settings
	// specified at the DestinationRule level.
	TrafficPolicy *TrafficPolicy `protobuf:"bytes,3,opt,name=traffic_policy,json=trafficPolicy" json:"traffic_policy,omitempty"`
}

type TrafficPolicy struct {
	// Settings controlling the load balancer algorithms.
	LoadBalancer *LoadBalancerSettings `protobuf:"bytes,1,opt,name=load_balancer,json=loadBalancer" json:"load_balancer,omitempty"`
	// Settings controlling the volume of connections to an upstream service
	ConnectionPool *ConnectionPoolSettings `protobuf:"bytes,2,opt,name=connection_pool,json=connectionPool" json:"connection_pool,omitempty"`
	// Settings controlling eviction of unhealthy hosts from the load balancing pool
	OutlierDetection *OutlierDetection `protobuf:"bytes,3,opt,name=outlier_detection,json=outlierDetection" json:"outlier_detection,omitempty"`
	// TLS related settings for connections to the upstream service.
	Tls *TLSSettings `protobuf:"bytes,4,opt,name=tls" json:"tls,omitempty"`
	// Traffic policies specific to individual ports. Note that port level
	// settings will override the destination-level settings. Traffic
	// settings specified at the destination-level will not be inherited when
	// overridden by port-level settings, i.e. default values will be applied
	// to fields omitted in port-level traffic policies.
	PortLevelSettings []*TrafficPolicy_PortTrafficPolicy `protobuf:"bytes,5,rep,name=port_level_settings,json=portLevelSettings" json:"port_level_settings,omitempty"`
}

type TrafficPolicy_PortTrafficPolicy struct {
	// Specifies the port name or number of a port on the destination service
	// on which this policy is being applied.
	//
	// Names must comply with DNS label syntax (rfc1035) and therefore cannot
	// collide with numbers. If there are multiple ports on a service with
	// the same protocol the names should be of the form <protocol-name>-<DNS
	// label>.
	Port *PortSelector `protobuf:"bytes,1,opt,name=port" json:"port,omitempty"`
	// Settings controlling the load balancer algorithms.
	LoadBalancer *LoadBalancerSettings `protobuf:"bytes,2,opt,name=load_balancer,json=loadBalancer" json:"load_balancer,omitempty"`
	// Settings controlling the volume of connections to an upstream service
	ConnectionPool *ConnectionPoolSettings `protobuf:"bytes,3,opt,name=connection_pool,json=connectionPool" json:"connection_pool,omitempty"`
	// Settings controlling eviction of unhealthy hosts from the load balancing pool
	OutlierDetection *OutlierDetection `protobuf:"bytes,4,opt,name=outlier_detection,json=outlierDetection" json:"outlier_detection,omitempty"`
	// TLS related settings for connections to the upstream service.
	Tls *TLSSettings `protobuf:"bytes,5,opt,name=tls" json:"tls,omitempty"`
}

type TLSSettings struct {
	// REQUIRED: Indicates whether connections to this port should be secured
	// using TLS. The value of this field determines how TLS is enforced.
	Mode int32 `protobuf:"varint,1,opt,name=mode,proto3,enum=istio.networking.v1alpha3.TLSSettings_TLSmode" json:"mode,omitempty"`
	// REQUIRED if mode is `MUTUAL`. The path to the file holding the
	// client-side TLS certificate to use.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	ClientCertificate string `protobuf:"bytes,2,opt,name=client_certificate,json=clientCertificate,proto3" json:"client_certificate,omitempty"`
	// REQUIRED if mode is `MUTUAL`. The path to the file holding the
	// client's private key.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	PrivateKey string `protobuf:"bytes,3,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	// OPTIONAL: The path to the file containing certificate authority
	// certificates to use in verifying a presented server certificate. If
	// omitted, the proxy will not verify the server's certificate.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	CaCertificates string `protobuf:"bytes,4,opt,name=ca_certificates,json=caCertificates,proto3" json:"ca_certificates,omitempty"`
	// A list of alternate names to verify the subject identity in the
	// certificate. If specified, the proxy will verify that the server
	// certificate's subject alt name matches one of the specified values.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	SubjectAltNames []string `protobuf:"bytes,5,rep,name=subject_alt_names,json=subjectAltNames" json:"subject_alt_names,omitempty"`
	// SNI string to present to the server during TLS handshake.
	// Should be empty if mode is `ISTIO_MUTUAL`.
	Sni string `protobuf:"bytes,6,opt,name=sni,proto3" json:"sni,omitempty"`
}

type OutlierDetection struct {
	// Number of errors before a host is ejected from the connection
	// pool. Defaults to 5. When the upstream host is accessed over HTTP, a
	// 5xx return code qualifies as an error. When the upstream host is
	// accessed over an opaque TCP connection, connect timeouts and
	// connection error/failure events qualify as an error.
	ConsecutiveErrors int32 `protobuf:"varint,1,opt,name=consecutive_errors,json=consecutiveErrors,proto3" json:"consecutive_errors,omitempty"`
	// Time interval between ejection sweep analysis. format:
	// 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s.
	Interval string `protobuf:"bytes,2,opt,name=interval" json:"interval,omitempty"`
	// Minimum ejection duration. A host will remain ejected for a period
	// equal to the product of minimum ejection duration and the number of
	// times the host has been ejected. This technique allows the system to
	// automatically increase the ejection period for unhealthy upstream
	// servers. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 30s.
	BaseEjectionTime string `protobuf:"bytes,3,opt,name=base_ejection_time,json=baseEjectionTime" json:"base_ejection_time,omitempty"`
	// Maximum % of hosts in the load balancing pool for the upstream
	// service that can be ejected. Defaults to 10%.
	MaxEjectionPercent int32 `protobuf:"varint,4,opt,name=max_ejection_percent,json=maxEjectionPercent,proto3" json:"max_ejection_percent,omitempty"`
}

type ConnectionPoolSettings struct {
	// Settings common to both HTTP and TCP upstream connections.
	Tcp *ConnectionPoolSettings_TCPSettings `protobuf:"bytes,1,opt,name=tcp" json:"tcp,omitempty"`
	// HTTP connection pool settings.
	Http *ConnectionPoolSettings_HTTPSettings `protobuf:"bytes,2,opt,name=http" json:"http,omitempty"`
}

type ConnectionPoolSettings_TCPSettings struct {
	// Maximum number of HTTP1 /TCP connections to a destination host.
	MaxConnections int32 `protobuf:"varint,1,opt,name=max_connections,json=maxConnections,proto3" json:"max_connections,omitempty"`
	// TCP connection timeout.
	ConnectTimeout string `protobuf:"bytes,2,opt,name=connect_timeout,json=connectTimeout" json:"connect_timeout,omitempty"`
}

type ConnectionPoolSettings_HTTPSettings struct {
	// Maximum number of pending HTTP requests to a destination. Default 1024.
	Http1MaxPendingRequests int32 `protobuf:"varint,1,opt,name=http1_max_pending_requests,json=http1MaxPendingRequests,proto3" json:"http1_max_pending_requests,omitempty"`
	// Maximum number of requests to a backend. Default 1024.
	Http2MaxRequests int32 `protobuf:"varint,2,opt,name=http2_max_requests,json=http2MaxRequests,proto3" json:"http2_max_requests,omitempty"`
	// Maximum number of requests per connection to a backend. Setting this
	// parameter to 1 disables keep alive.
	MaxRequestsPerConnection int32 `protobuf:"varint,3,opt,name=max_requests_per_connection,json=maxRequestsPerConnection,proto3" json:"max_requests_per_connection,omitempty"`
	// Maximum number of retries that can be outstanding to all hosts in a
	// cluster at a given time. Defaults to 3.
	MaxRetries int32 `protobuf:"varint,4,opt,name=max_retries,json=maxRetries,proto3" json:"max_retries,omitempty"`
}

type LoadBalancerSettings struct {
	// Upstream load balancing policy.
	//
	// Types that are valid to be assigned to LbPolicy:
	//	*LoadBalancerSettings_Simple
	//	*LoadBalancerSettings_ConsistentHash
	Simple         int32                                  `json:"simple,omitempty"`
	ConsistentHash *LoadBalancerSettings_ConsistentHashLB `json:"consistentHash,omitempty"`
}

type LoadBalancerSettings_ConsistentHashLB struct {
	// REQUIRED: The hash key to use.
	//
	// Types that are valid to be assigned to HashKey:
	//	*LoadBalancerSettings_ConsistentHashLB_HttpHeaderName
	//	*LoadBalancerSettings_ConsistentHashLB_HttpCookie
	//	*LoadBalancerSettings_ConsistentHashLB_UseSourceIp
	HttpHeaderName string                                            `json:"httpHeaderName,omitempty"`
	HttpCookie     *LoadBalancerSettings_ConsistentHashLB_HTTPCookie `json:"httpCookie,omitempty"`
	UseSourceIp    bool                                              `json:"useSourceIp,omitempty"`

	// The minimum number of virtual nodes to use for the hash
	// ring. Defaults to 1024. Larger ring sizes result in more granular
	// load distributions. If the number of hosts in the load balancing
	// pool is larger than the ring size, each host will be assigned a
	// single virtual node.
	MinimumRingSize uint64 `protobuf:"varint,4,opt,name=minimum_ring_size,json=minimumRingSize,proto3" json:"minimum_ring_size,omitempty"`
}

type LoadBalancerSettings_ConsistentHashLB_HTTPCookie struct {
	// REQUIRED. Name of the cookie.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Path to set for the cookie.
	Path string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	// REQUIRED. Lifetime of the cookie.
	Ttl *time.Duration `protobuf:"bytes,3,opt,name=ttl,stdduration" json:"ttl,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DestinationRuleList is a list of DestinationRule resources
type DestinationRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DestinationRule `json:"items"`
}
