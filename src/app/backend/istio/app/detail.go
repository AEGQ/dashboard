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
	api2 "github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/namespace"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetAppDeploySpec queries the specified application's k8s deployment.
func GetAppDeploySpec(client kubernetes.Interface, namespace *common.NamespaceQuery, appName string) ([]v1beta1.Deployment, error) {
	return getDeploymentByLabels(client, namespace, map[string]string{
		"app": appName,
	})
}

// GetAppDetail queries the specified application's deploy.
func GetAppDetail(client kubernetes.Interface, istioClient istio.Interface, nsQuery *common.NamespaceQuery, appName string, dataSelect *dataselect.DataSelectQuery) (*api.App, error) {
	channels := common.ResourceChannels{
		DestinationRuleList: common.GetDestinationRuleListChannel(istioClient, nsQuery, 1),
	}

	svc, err := client.CoreV1().Services(nsQuery.ToRequestParam()).Get(appName, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	vServices, err := virtualservice.GetVirtualServices(istioClient, []string{virtualservice.FQDN(appName, nsQuery.ToRequestParam())}, virtualservice.All)
	if err != nil {
		return nil, err
	}

	ns, err := namespace.GetNamespaceDetail(client, nsQuery.ToRequestParam())
	if err != nil {
		return nil, err
	}

	dRules := <-channels.DestinationRuleList.List
	err = <-channels.DestinationRuleList.Error
	if err != nil {
		return nil, err
	}

	// merge destinationRules & services
	var app = getAppDetail(svc, vServices, dRules.Items, ns)
	app.Metrics = *GetAppMetrics(client, app)
	return app, nil
}

func getAppDetail(svc *v1.Service, vServices []istioApi.VirtualService,
	dRules []istioApi.DestinationRule, namespace *namespace.NamespaceDetail) *api.App {
	app := &api.App{
		ObjectMeta: api2.NewObjectMeta(svc.ObjectMeta),
		TypeMeta: api2.TypeMeta{
			Kind: api2.ResourceKindApp,
		},
		Destinations: []api.Destination{
			{
				Version:  "default",
				Selector: svc.Spec.Selector,
			},
		},
	}

	if dRules != nil && len(dRules) > 0 {
		app.Destinations = getAppDestinations(app, dRules)
	}

	if vServices != nil && len(vServices) > 0 {
		app.VirtualServices = getAppVirtualServices(app, vServices)
	}

	// add these applications' istio statuses
	if namespace != nil {
		if namespace.ObjectMeta.Labels["istio-injection"] == "enabled" {
			app.Status.Istio = true
		}
	}
	return app
}

// getDeploymentByLabels filters deployments by the labels.
func getDeploymentByLabels(client kubernetes.Interface, namespace *common.NamespaceQuery, filterLabels map[string]string) ([]v1beta1.Deployment, error) {
	deployments, err := client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var deps []v1beta1.Deployment
	for _, dep := range deployments.Items {
		matched := true
		for key, val := range filterLabels {
			if dep.ObjectMeta.Labels[key] != val {
				matched = false
				break
			}
		}
		if matched {
			deps = append(deps, dep)
		}
	}
	return deps, nil
}
