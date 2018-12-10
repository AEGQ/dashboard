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

package api

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	v1 "k8s.io/api/core/v1"
)

type AppList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	Apps []*App `json:"apps"`
}

type App struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	Status          Status                    `json:"status"`
	VirtualServices []v1alpha3.VirtualService `json:"virtualServices,omitempty"`
	Destinations    []Destination             `json:"destinations"`
	Metrics         Metrics                   `json:"metrics,omitempty"`
}

// Metrics is Istio application metrics collected by Prometheus which is is the default
// Grafana Dashboard provided by Istio, Metrics configures 4 metric panel urls.
// for example:
// http://grafana/d-solo/LJ_uJAvmk/istio-service-dashboard?refresh=10s&orgId=1&panelId=25&var-service={{appName}}&var-srcns=All&var-srcwl=All&var-dstns=All&var-dstwl=All
// See: https://istio.io/docs/tasks/telemetry/using-istio-dashboard/
type Metrics struct {
	ClientQps     string `json:"clientQps,omitempty"`
	ClientLatency string `json:"clientLatency,omitempty"`
	ServerQps     string `json:"serverQps,omitempty"`
	ServerLatency string `json:"serverLatency,omitempty"`
}

type Status struct {
	// Istio indicates if this application is istio-enabled
	Istio bool `json:"istio"`
}

type Destination struct {
	Version  string            `json:"version"`
	Selector map[string]string `json:"selector"`
}

type NewApplication struct {
	Version  string `json:"version"`
	Replicas int32  `json:"replicas"`

	Ports       []v1.ServicePort   `json:"ports"`
	PodTemplate v1.PodTemplateSpec `json:"podTemplate"`
}

type CanaryDeployment struct {
	Version  string `json:"version"`
	Replicas int32  `json:"replicas"`

	PodTemplate v1.PodTemplateSpec `json:"podTemplate"`
}

// IngressList indicates a list of istio ingresses.
type IngressList struct {
	api.ListMeta `json:"listMeta"`

	Items  []Ingress `json:"items"`
	Errors []error   `json:"errors"`
}

// Ingress represents the istio ingress including `Gateway`, `VirtualService`, `DestinationRule`,
// `Deployment`.
type Ingress struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`

	Hosts             []common.Endpoint         `json:"hosts,omitempty"`
	ExternalEndpoints []common.Endpoint         `json:"externalEndpoints"`
	VirtualServices   []v1alpha3.VirtualService `json:"virtualServices,omitempty"`

	Errors []error `json:"errors"`
}
