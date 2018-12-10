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
	"strings"

	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetAppMetrics fetches metric urls from k8s configMap and returns grafana metric embedded graphs.
// ```yaml
// kind: ConfigMap
// apiVersion: v1
// metadata:
//   name: istio-app-metrics
//   namespace: istio-system
// data:
//   clientQPS: http://grafana_url/d-solo/LJ_uJAvmk/istio-service-dashboard?refresh=10s&orgId=1&panelId=25&var-service={{AppName}}&var-srcns=All&var-srcwl=All&var-dstns=All&var-dstwl=All
//   clientLatency: http://grafana_url/d-solo/LJ_uJAvmk/istio-service-dashboard?refresh=10s&orgId=1&panelId=27&var-service={{AppName}}&var-srcns=All&var-srcwl=All&var-dstns=All&var-dstwl=All
//   serverQPS: http://grafana_url/d-solo/LJ_uJAvmk/istio-service-dashboard?refresh=10s&orgId=1&panelId=90&var-service={{AppName}}&var-srcns=All&var-srcwl=All&var-dstns=All&var-dstwl=All
//   serverLatency: http://grafana_url/d-solo/LJ_uJAvmk/istio-service-dashboard?refresh=10s&orgId=1&panelId=94&var-service={{AppName}}&var-srcns=All&var-srcwl=All&var-dstns=All&var-dstwl=All
// ```
func GetAppMetrics(client kubernetes.Interface, app *api.App) *api.Metrics {
	metric := &api.Metrics{}
	config, err := client.CoreV1().ConfigMaps("istio-system").
		Get("istio-app-metrics", metaV1.GetOptions{})
	if err != nil {
		return metric
	}

	appAddr := virtualservice.FQDN(app.ObjectMeta.Name, app.ObjectMeta.Namespace)
	metric.ClientQps = strings.Replace(config.Data["clientQPS"], "{{AppName}}", appAddr, -1)
	metric.ClientLatency = strings.Replace(config.Data["clientLatency"], "{{AppName}}", appAddr, -1)
	metric.ServerQps = strings.Replace(config.Data["serverQPS"], "{{AppName}}", appAddr, -1)
	metric.ServerLatency = strings.Replace(config.Data["serverLatency"], "{{AppName}}", appAddr, -1)
	return metric
}
