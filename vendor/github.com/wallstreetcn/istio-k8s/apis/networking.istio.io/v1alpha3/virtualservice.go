package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualService describes istio's flow control.
type VirtualService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualServiceSpec `json:"spec"`
}

type VirtualServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Hosts    []string     `protobuf:"bytes,1,rep,name=hosts" json:"hosts,omitempty"`
	Gateways []string     `protobuf:"bytes,2,rep,name=gateways" json:"gateways,omitempty"`
	Http     []*HTTPRoute `protobuf:"bytes,3,rep,name=http" json:"http,omitempty"`
	Tls      []*TLSRoute  `protobuf:"bytes,5,rep,name=tls" json:"tls,omitempty"`
	Tcp      []*TCPRoute  `protobuf:"bytes,4,rep,name=tcp" json:"tcp,omitempty"`
}

type StringMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

// HttpMatchRequest specifies a set of criterion to be met in order for the
// rule to be applied to the HTTP request. For example, the following
// restricts the rule to match only requests where the URL path
// starts with /ratings/v2/ and the request contains a custom `end-user` header
// with value `jason`.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: ratings-route
// spec:
//   hosts:
//   - ratings.prod.svc.cluster.local
//   http:
//   - match:
//     - headers:
//         end-user:
//           exact: jason
//       uri:
//         prefix: "/ratings/v2/"
//     route:
//     - destination:
//         host: ratings.prod.svc.cluster.local
// ```
//
// HTTPMatchRequest CANNOT be empty.
type HTTPMatchRequest struct {
	// URI to match
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Uri *StringMatch `protobuf:"bytes,1,opt,name=uri" json:"uri,omitempty"`
	// URI Scheme
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Scheme *StringMatch `protobuf:"bytes,2,opt,name=scheme" json:"scheme,omitempty"`
	// HTTP Method
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Method *StringMatch `protobuf:"bytes,3,opt,name=method" json:"method,omitempty"`
	// HTTP Authority
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Authority *StringMatch `protobuf:"bytes,4,opt,name=authority" json:"authority,omitempty"`
	// The header keys must be lowercase and use hyphen as the separator,
	// e.g. _x-request-id_.
	//
	// Header values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	// **Note:** The keys `uri`, `scheme`, `method`, and `authority` will be ignored.
	Headers map[string]*StringMatch `protobuf:"bytes,5,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value"`
	// Specifies the ports on the host that is being addressed. Many services
	// only expose a single port or label ports with the protocols they support,
	// in these cases it is not required to explicitly select the port.
	Port uint32 `protobuf:"varint,6,opt,name=port,proto3" json:"port,omitempty"`
	// One or more labels that constrain the applicability of a rule to
	// workloads with the given labels. If the VirtualService has a list of
	// gateways specified at the top, it should include the reserved gateway
	// `mesh` in order for this field to be applicable.
	SourceLabels map[string]string `protobuf:"bytes,7,rep,name=source_labels,json=sourceLabels" json:"source_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Names of gateways where the rule should be applied to. Gateway names
	// at the top of the VirtualService (if any) are overridden. The gateway match is
	// independent of sourceLabels.
	Gateways []string `protobuf:"bytes,8,rep,name=gateways" json:"gateways,omitempty"`
}

// Describes match conditions and actions for routing HTTP/1.1, HTTP2, and
// gRPC traffic. See VirtualService for usage examples.
type HTTPRoute struct {
	// Match conditions to be satisfied for the rule to be
	// activated. All conditions inside a single match block have AND
	// semantics, while the list of match blocks have OR semantics. The rule
	// is matched if any one of the match blocks succeed.
	Match []*HTTPMatchRequest `protobuf:"bytes,1,rep,name=match" json:"match,omitempty"`
	// A http rule can either redirect or forward (default) traffic. The
	// forwarding target can be one of several versions of a service (see
	// glossary in beginning of document). Weights associated with the
	// service version determine the proportion of traffic it receives.
	Route []*DestinationWeight `protobuf:"bytes,2,rep,name=route" json:"route,omitempty"`
	// A http rule can either redirect or forward (default) traffic. If
	// traffic passthrough option is specified in the rule,
	// route/redirect will be ignored. The redirect primitive can be used to
	// send a HTTP 301 redirect to a different URI or Authority.
	Redirect *HTTPRedirect `protobuf:"bytes,3,opt,name=redirect" json:"redirect,omitempty"`
	// Rewrite HTTP URIs and Authority headers. Rewrite cannot be used with
	// Redirect primitive. Rewrite will be performed before forwarding.
	Rewrite *HTTPRewrite `protobuf:"bytes,4,opt,name=rewrite" json:"rewrite,omitempty"`
	// Deprecated. Websocket upgrades are done automatically starting from Istio 1.0.
	// $hide_from_docs
	WebsocketUpgrade bool `protobuf:"varint,5,opt,name=websocket_upgrade,json=websocketUpgrade,proto3" json:"websocket_upgrade,omitempty"`
	// Timeout for HTTP requests.
	Timeout string `protobuf:"bytes,6,opt,name=timeout" json:"timeout,omitempty"`
	// Retry policy for HTTP requests.
	Retries *HTTPRetry `protobuf:"bytes,7,opt,name=retries" json:"retries,omitempty"`
	// Fault injection policy to apply on HTTP traffic at the client side.
	// Note that timeouts or retries will not be enabled when faults are
	// enabled on the client side.
	Fault *HTTPFaultInjection `protobuf:"bytes,8,opt,name=fault" json:"fault,omitempty"`
	// Mirror HTTP traffic to a another destination in addition to forwarding
	// the requests to the intended destination. Mirrored traffic is on a
	// best effort basis where the sidecar/gateway will not wait for the
	// mirrored cluster to respond before returning the response from the
	// original destination.  Statistics will be generated for the mirrored
	// destination.
	Mirror *Destination `protobuf:"bytes,9,opt,name=mirror" json:"mirror,omitempty"`
	// Cross-Origin Resource Sharing policy (CORS). Refer to
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS
	// for further details about cross origin resource sharing.
	CorsPolicy *CorsPolicy `protobuf:"bytes,10,opt,name=cors_policy,json=corsPolicy" json:"cors_policy,omitempty"`
	// Additional HTTP headers to add before forwarding a request to the
	// destination service.
	AppendHeaders map[string]string `protobuf:"bytes,11,rep,name=append_headers,json=appendHeaders" json:"append_headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Http headers to remove before returning the response to the caller
	// $hide_from_docs
	RemoveResponseHeaders []string `protobuf:"bytes,12,rep,name=remove_response_headers,json=removeResponseHeaders" json:"remove_response_headers,omitempty"`
}

type CorsPolicy struct {
	// The list of origins that are allowed to perform CORS requests. The
	// content will be serialized into the Access-Control-Allow-Origin
	// header. Wildcard * will allow all origins.
	AllowOrigin []string `protobuf:"bytes,1,rep,name=allow_origin,json=allowOrigin" json:"allow_origin,omitempty"`
	// List of HTTP methods allowed to access the resource. The content will
	// be serialized into the Access-Control-Allow-Methods header.
	AllowMethods []string `protobuf:"bytes,2,rep,name=allow_methods,json=allowMethods" json:"allow_methods,omitempty"`
	// List of HTTP headers that can be used when requesting the
	// resource. Serialized to Access-Control-Allow-Methods header.
	AllowHeaders []string `protobuf:"bytes,3,rep,name=allow_headers,json=allowHeaders" json:"allow_headers,omitempty"`
	// A white list of HTTP headers that the browsers are allowed to
	// access. Serialized into Access-Control-Expose-Headers header.
	ExposeHeaders []string `protobuf:"bytes,4,rep,name=expose_headers,json=exposeHeaders" json:"expose_headers,omitempty"`
	// Specifies how long the the results of a preflight request can be
	// cached. Translates to the Access-Control-Max-Age header.
	MaxAge string `protobuf:"bytes,5,opt,name=max_age,json=maxAge" json:"max_age,omitempty"`
	// Indicates whether the caller is allowed to send the actual request
	// (not the preflight) using credentials. Translates to
	// Access-Control-Allow-Credentials header.
	AllowCredentials bool `protobuf:"bytes,6,opt,name=allow_credentials,json=allowCredentials" json:"allow_credentials,omitempty"`
}

type HTTPFaultInjection struct {
	// Delay requests before forwarding, emulating various failures such as
	// network issues, overloaded upstream service, etc.
	Delay *HTTPFaultInjection_Delay `protobuf:"bytes,1,opt,name=delay" json:"delay,omitempty"`
	// Abort Http request attempts and return error codes back to downstream
	// service, giving the impression that the upstream service is faulty.
	Abort *HTTPFaultInjection_Abort `protobuf:"bytes,2,opt,name=abort" json:"abort,omitempty"`
}

type HTTPFaultInjection_Abort struct {
	// Percentage of requests to be aborted with the error code provided (0-100).
	Percent int32 `protobuf:"varint,1,opt,name=percent,proto3" json:"percent,omitempty"`
	// Types that are valid to be assigned to ErrorType:
	//	*HTTPFaultInjection_Abort_HttpStatus
	//	*HTTPFaultInjection_Abort_GrpcStatus
	//	*HTTPFaultInjection_Abort_Http2Error
	HttpStatus int `json:"httpStatus,omitEmpty"`
	GrpcStatus int `json:"grpcStatus,omitEmpty"`
	Http2Error int `json:"http2Error,omitEmpty"`
}

type HTTPFaultInjection_Delay struct {
	// Percentage of requests on which the delay will be injected (0-100).
	Percent int32 `protobuf:"varint,1,opt,name=percent,proto3" json:"percent,omitempty"`
	// Types that are valid to be assigned to HttpDelayType:
	//	*HTTPFaultInjection_Delay_FixedDelay
	//	*HTTPFaultInjection_Delay_ExponentialDelay
	FixedDelay       string `json:"fixedDelay,omitempty"`
	ExponentialDelay string `json:"exponential_delay,omitempty"`
}

type HTTPRetry struct {
	// REQUIRED. Number of retries for a given request. The interval
	// between retries will be determined automatically (25ms+). Actual
	// number of retries attempted depends on the httpReqTimeout.
	Attempts int32 `protobuf:"varint,1,opt,name=attempts,proto3" json:"attempts,omitempty"`
	// Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms.
	PerTryTimeout string `protobuf:"bytes,2,opt,name=per_try_timeout,json=perTryTimeout" json:"per_try_timeout,omitempty"`
}

// HTTPRewrite can be used to rewrite specific parts of a HTTP request
// before forwarding the request to the destination. Rewrite primitive can
// be used only with the DestinationWeights. The following example
// demonstrates how to rewrite the URL prefix for api call (/ratings) to
// ratings service before making the actual API call.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: ratings-route
// spec:
//   hosts:
//   - ratings.prod.svc.cluster.local
//   http:
//   - match:
//     - uri:
//         prefix: /ratings
//     rewrite:
//       uri: /v1/bookRatings
//     route:
//     - destination:
//         host: ratings.prod.svc.cluster.local
//         subset: v1
// ```
//
type HTTPRewrite struct {
	// rewrite the path (or the prefix) portion of the URI with this
	// value. If the original URI was matched based on prefix, the value
	// provided in this field will replace the corresponding matched prefix.
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	// rewrite the Authority/Host header with this value.
	Authority string `protobuf:"bytes,2,opt,name=authority,proto3" json:"authority,omitempty"`
}

// HTTPRedirect can be used to send a 301 redirect response to the caller,
// where the Authority/Host and the URI in the response can be swapped with
// the specified values. For example, the following rule redirects
// requests for /v1/getProductRatings API on the ratings service to
// /v1/bookRatings provided by the bookratings service.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: ratings-route
// spec:
//   hosts:
//   - ratings.prod.svc.cluster.local
//   http:
//   - match:
//     - uri:
//         exact: /v1/getProductRatings
//   redirect:
//     uri: /v1/bookRatings
//     authority: newratings.default.svc.cluster.local
//   ...
// ```
type HTTPRedirect struct {
	// On a redirect, overwrite the Path portion of the URL with this
	// value. Note that the entire path will be replaced, irrespective of the
	// request URI being matched as an exact path or prefix.
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	// On a redirect, overwrite the Authority/Host portion of the URL with
	// this value.
	Authority string `protobuf:"bytes,2,opt,name=authority,proto3" json:"authority,omitempty"`
}

// Each routing rule is associated with one or more service versions (see
// glossary in beginning of document). Weights associated with the version
// determine the proportion of traffic it receives. For example, the
// following rule will route 25% of traffic for the "reviews" service to
// instances with the "v2" tag and the remaining traffic (i.e., 75%) to
// "v1".
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: reviews-route
// spec:
//   hosts:
//   - reviews.prod.svc.cluster.local
//   http:
//   - route:
//     - destination:
//         host: reviews.prod.svc.cluster.local
//         subset: v2
//       weight: 25
//     - destination:
//         host: reviews.prod.svc.cluster.local
//         subset: v1
//       weight: 75
// ```
//
// And the associated DestinationRule
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: reviews-destination
// spec:
//   host: reviews.prod.svc.cluster.local
//   subsets:
//   - name: v1
//     labels:
//       version: v1
//   - name: v2
//     labels:
//       version: v2
// ```
//
// Traffic can also be split across two entirely different services without
// having to define new subsets. For example, the following rule forwards 25% of
// traffic to reviews.com to dev.reviews.com
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: reviews-route-two-domains
// spec:
//   hosts:
//   - reviews.com
//   http:
//   - route:
//     - destination:
//         host: dev.reviews.com
//       weight: 25
//     - destination:
//         host: reviews.com
//       weight: 75
// ```
type DestinationWeight struct {
	// REQUIRED. Destination uniquely identifies the instances of a service
	// to which the request/connection should be forwarded to.
	Destination *Destination `protobuf:"bytes,1,opt,name=destination" json:"destination,omitempty"`
	// REQUIRED. The proportion of traffic to be forwarded to the service
	// version. (0-100). Sum of weights across destinations SHOULD BE == 100.
	// If there is only destination in a rule, the weight value is assumed to
	// be 100.
	Weight int32 `protobuf:"varint,2,opt,name=weight,proto3" json:"weight,omitempty"`
}

// Destination indicates the network addressable service to which the
// request/connection will be sent after processing a routing rule. The
// destination.host should unambiguously refer to a service in the service
// registry. Istio's service registry is composed of all the services found
// in the platform's service registry (e.g., Kubernetes services, Consul
// services), as well as services declared through the
// [ServiceEntry](#ServiceEntry) resource.
//
// *Note for Kubernetes users*: When short names are used (e.g. "reviews"
// instead of "reviews.default.svc.cluster.local"), Istio will interpret
// the short name based on the namespace of the rule, not the service. A
// rule in the "default" namespace containing a host "reviews will be
// interpreted as "reviews.default.svc.cluster.local", irrespective of the
// actual namespace associated with the reviews service. _To avoid potential
// misconfigurations, it is recommended to always use fully qualified
// domain names over short names._
//
// The following Kubernetes example routes all traffic by default to pods
// of the reviews service with label "version: v1" (i.e., subset v1), and
// some to subset v2, in a kubernetes environment.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: reviews-route
//   namespace: foo
// spec:
//   hosts:
//   - reviews # interpreted as reviews.foo.svc.cluster.local
//   http:
//   - match:
//     - uri:
//         prefix: "/wpcatalog"
//     - uri:
//         prefix: "/consumercatalog"
//     rewrite:
//       uri: "/newcatalog"
//     route:
//     - destination:
//         host: reviews # interpreted as reviews.foo.svc.cluster.local
//         subset: v2
//   - route:
//     - destination:
//         host: reviews # interpreted as reviews.foo.svc.cluster.local
//         subset: v1
// ```
//
// And the associated DestinationRule
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: reviews-destination
//   namespace: foo
// spec:
//   host: reviews # interpreted as reviews.foo.svc.cluster.local
//   subsets:
//   - name: v1
//     labels:
//       version: v1
//   - name: v2
//     labels:
//       version: v2
// ```
//
// The following VirtualService sets a timeout of 5s for all calls to
// productpage.prod.svc.cluster.local service in Kubernetes. Notice that
// there are no subsets defined in this rule. Istio will fetch all
// instances of productpage.prod.svc.cluster.local service from the service
// registry and populate the sidecar's load balancing pool. Also, notice
// that this rule is set in the istio-system namespace but uses the fully
// qualified domain name of the productpage service,
// productpage.prod.svc.cluster.local. Therefore the rule's namespace does
// not have an impact in resolving the name of the productpage service.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: my-productpage-rule
//   namespace: istio-system
// spec:
//   hosts:
//   - productpage.prod.svc.cluster.local # ignores rule namespace
//   http:
//   - timeout: 5s
//     route:
//     - destination:
//         host: productpage.prod.svc.cluster.local
// ```
//
// To control routing for traffic bound to services outside the mesh, external
// services must first be added to Istio's internal service registry using the
// ServiceEntry resource. VirtualServices can then be defined to control traffic
// bound to these external services. For example, the following rules define a
// Service for wikipedia.org and set a timeout of 5s for http requests.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: ServiceEntry
// metadata:
//   name: external-svc-wikipedia
// spec:
//   hosts:
//   - wikipedia.org
//   location: MESH_EXTERNAL
//   ports:
//   - number: 80
//     name: example-http
//     protocol: HTTP
//   resolution: DNS
//
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: my-wiki-rule
// spec:
//   hosts:
//   - wikipedia.org
//   http:
//   - timeout: 5s
//     route:
//     - destination:
//         host: wikipedia.org
// ```
type Destination struct {
	// REQUIRED. The name of a service from the service registry. Service
	// names are looked up from the platform's service registry (e.g.,
	// Kubernetes services, Consul services, etc.) and from the hosts
	// declared by [ServiceEntry](#ServiceEntry). Traffic forwarded to
	// destinations that are not found in either of the two, will be dropped.
	//
	// *Note for Kubernetes users*: When short names are used (e.g. "reviews"
	// instead of "reviews.default.svc.cluster.local"), Istio will interpret
	// the short name based on the namespace of the rule, not the service. A
	// rule in the "default" namespace containing a host "reviews will be
	// interpreted as "reviews.default.svc.cluster.local", irrespective of
	// the actual namespace associated with the reviews service. _To avoid
	// potential misconfigurations, it is recommended to always use fully
	// qualified domain names over short names._
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// The name of a subset within the service. Applicable only to services
	// within the mesh. The subset must be defined in a corresponding
	// DestinationRule.
	Subset string `protobuf:"bytes,2,opt,name=subset,proto3" json:"subset,omitempty"`
	// Specifies the port on the host that is being addressed. If a service
	// exposes only a single port it is not required to explicitly select the
	// port.
	Port *PortSelector `protobuf:"bytes,3,opt,name=port" json:"port,omitempty"`
}

// PortSelector specifies the number of a port to be used for
// matching or selection for final routing.
type PortSelector struct {
	// Types that are valid to be assigned to Port:
	//	*PortSelector_Number
	//	*PortSelector_Name
	Number uint32 `json:"number,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Describes match conditions and actions for routing unterminated TLS
// traffic (TLS/HTTPS) The following routing rule forwards unterminated TLS
// traffic arriving at port 443 of gateway called "mygateway" to internal
// services in the mesh based on the SNI value.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: bookinfo-sni
// spec:
//   hosts:
//   - "*.bookinfo.com"
//   gateways:
//   - mygateway
//   tls:
//   - match:
//     - port: 443
//       sniHosts:
//       - login.bookinfo.com
//     route:
//     - destination:
//         host: login.prod.svc.cluster.local
//   - match:
//     - port: 443
//       sniHosts:
//       - reviews.bookinfo.com
//     route:
//     - destination:
//         host: reviews.prod.svc.cluster.local
// ```
type TLSRoute struct {
	// REQUIRED. Match conditions to be satisfied for the rule to be
	// activated. All conditions inside a single match block have AND
	// semantics, while the list of match blocks have OR semantics. The rule
	// is matched if any one of the match blocks succeed.
	Match []*TLSMatchAttributes `protobuf:"bytes,1,rep,name=match" json:"match,omitempty"`
	// The destination to which the connection should be forwarded to.
	// Currently, only one destination is allowed for TLS services. When TCP
	// weighted routing support is introduced in Envoy, multiple destinations
	// with weights can be specified.
	Route []*DestinationWeight `protobuf:"bytes,2,rep,name=route" json:"route,omitempty"`
}

// TLS connection match attributes.
type TLSMatchAttributes struct {
	// REQUIRED. SNI (server name indicator) to match on. Wildcard prefixes
	// can be used in the SNI value. E.g., *.com will match foo.example.com
	// as well as example.com.
	SniHosts []string `protobuf:"bytes,1,rep,name=sni_hosts,json=sniHosts" json:"sni_hosts,omitempty"`
	// IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g.,
	// a.b.c.d/xx form or just a.b.c.d.
	DestinationSubnets []string `protobuf:"bytes,2,rep,name=destination_subnets,json=destinationSubnets" json:"destination_subnets,omitempty"`
	// Specifies the port on the host that is being addressed. Many services
	// only expose a single port or label ports with the protocols they
	// support, in these cases it is not required to explicitly select the
	// port.
	Port uint32 `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	// IPv4 or IPv6 ip address of source with optional subnet. E.g., a.b.c.d/xx
	// form or just a.b.c.d
	// $hide_from_docs
	SourceSubnet string `protobuf:"bytes,4,opt,name=source_subnet,json=sourceSubnet,proto3" json:"source_subnet,omitempty"`
	// One or more labels that constrain the applicability of a rule to
	// workloads with the given labels. If the VirtualService has a list of
	// gateways specified at the top, it should include the reserved gateway
	// `mesh` in order for this field to be applicable.
	SourceLabels map[string]string `protobuf:"bytes,5,rep,name=source_labels,json=sourceLabels" json:"source_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Names of gateways where the rule should be applied to. Gateway names
	// at the top of the VirtualService (if any) are overridden. The gateway
	// match is independent of sourceLabels.
	Gateways []string `protobuf:"bytes,6,rep,name=gateways" json:"gateways,omitempty"`
}

// Describes match conditions and actions for routing TCP traffic. The
// following routing rule forwards traffic arriving at port 27017 for
// mongo.prod.svc.cluster.local to another Mongo server on port 5555.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: bookinfo-Mongo
// spec:
//   hosts:
//   - mongo.prod.svc.cluster.local
//   tcp:
//   - match:
//     - port: 27017
//     route:
//     - destination:
//         host: mongo.backup.svc.cluster.local
//         port:
//           number: 5555
// ```
type TCPRoute struct {
	// Match conditions to be satisfied for the rule to be
	// activated. All conditions inside a single match block have AND
	// semantics, while the list of match blocks have OR semantics. The rule
	// is matched if any one of the match blocks succeed.
	Match []*L4MatchAttributes `protobuf:"bytes,1,rep,name=match" json:"match,omitempty"`
	// The destination to which the connection should be forwarded to.
	// Currently, only one destination is allowed for TCP services. When TCP
	// weighted routing support is introduced in Envoy, multiple destinations
	// with weights can be specified.
	Route []*DestinationWeight `protobuf:"bytes,2,rep,name=route" json:"route,omitempty"`
}

// L4 connection match attributes. Note that L4 connection matching support
// is incomplete.
type L4MatchAttributes struct {
	// IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g.,
	// a.b.c.d/xx form or just a.b.c.d.
	DestinationSubnets []string `protobuf:"bytes,1,rep,name=destination_subnets,json=destinationSubnets" json:"destination_subnets,omitempty"`
	// Specifies the port on the host that is being addressed. Many services
	// only expose a single port or label ports with the protocols they support,
	// in these cases it is not required to explicitly select the port.
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// IPv4 or IPv6 ip address of source with optional subnet. E.g., a.b.c.d/xx
	// form or just a.b.c.d
	// $hide_from_docs
	SourceSubnet string `protobuf:"bytes,3,opt,name=source_subnet,json=sourceSubnet,proto3" json:"source_subnet,omitempty"`
	// One or more labels that constrain the applicability of a rule to
	// workloads with the given labels. If the VirtualService has a list of
	// gateways specified at the top, it should include the reserved gateway
	// `mesh` in order for this field to be applicable.
	SourceLabels map[string]string `protobuf:"bytes,4,rep,name=source_labels,json=sourceLabels" json:"source_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Names of gateways where the rule should be applied to. Gateway names
	// at the top of the VirtualService (if any) are overridden. The gateway
	// match is independent of sourceLabels.
	Gateways []string `protobuf:"bytes,5,rep,name=gateways" json:"gateways,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VirtualServiceList is a list of VirtualService resources
type VirtualServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VirtualService `json:"items"`
}
