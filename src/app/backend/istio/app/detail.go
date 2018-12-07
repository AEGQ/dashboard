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
	"fmt"

	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/destinationrule"
	"github.com/kubernetes/dashboard/src/app/backend/resource/namespace"
	"github.com/kubernetes/dashboard/src/app/backend/resource/service"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
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
func GetAppDetail(client kubernetes.Interface, istioClient istio.Interface, ns *common.NamespaceQuery, appName string, dataSelect *dataselect.DataSelectQuery) (*api.App, error) {
	dRules, err := destinationrule.GetDestinationRuleList(client, istioClient, ns, dataselect.NoDataSelect)
	if err != nil {
		return nil, err
	}

	services, err := service.GetServiceList(client, ns, dataSelect)
	if err != nil {
		return nil, err
	}

	// TODO to be refactored
	svcs := []service.Service{}
	for _, svc := range services.Services {
		if svc.ObjectMeta.Name == appName {
			svcs = append(svcs, svc)
		}
	}
	services.Services = svcs
	services.ListMeta.TotalItems = len(svcs)

	vServices, err := virtualservice.GetVirtualServices(istioClient, []string{virtualservice.FQDN(appName, ns.ToRequestParam())}, virtualservice.All)
	if err != nil {
		return nil, err
	}

	namespaces, err := namespace.GetNamespaceList(client, dataselect.NoDataSelect)
	if err != nil {
		return nil, err
	}

	// merge destinationRules & services
	result, err := getApps(services, dRules, vServices, namespaces)
	if err != nil {
		return nil, err
	}

	if result.ListMeta.TotalItems != 1 {
		return nil, fmt.Errorf("one application expected, %d found", result.ListMeta.TotalItems)
	}
	return result.Apps[0], nil
}

// getDeploymentByLabels filters deployments by the labels.
func getDeploymentByLabels(client kubernetes.Interface, namespace *common.NamespaceQuery, filterLabels map[string]string) ([]v1beta1.Deployment, error) {
	deployments, err := client.ExtensionsV1beta1().Deployments(namespace.ToRequestParam()).List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var parentDeps []v1beta1.Deployment
	for _, dep := range deployments.Items {
		matched := true
		for key, val := range filterLabels {
			if dep.ObjectMeta.Labels[key] != val {
				matched = false
				break
			}
		}
		if matched {
			parentDeps = append(parentDeps, dep)
		}
	}
	return parentDeps, nil
}
