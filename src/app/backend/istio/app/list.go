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
	"errors"

	api2 "github.com/kubernetes/dashboard/src/app/backend/api"
	kdErrors "github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/service"
	istioApi "github.com/wallstreetcn/istio-k8s/apis/networking.istio.io/v1alpha3"
	istio "github.com/wallstreetcn/istio-k8s/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetAppList queries istio app list.
func GetAppList(client kubernetes.Interface, istioClient istio.Interface, ns *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*api.AppList, error) {
	channels := &common.ResourceChannels{
		ServiceList:         common.GetServiceListChannel(client, ns, 1),
		NamespaceList:       common.GetNamespaceListChannel(client, 1),
		DestinationRuleList: common.GetDestinationRuleListChannel(istioClient, ns, 1),
	}
	return getAppsFromChannel(channels, dsQuery)
}

func getAppsFromChannel(channels *common.ResourceChannels, dsQuery *dataselect.DataSelectQuery) (*api.AppList, error) {
	services := <-channels.ServiceList.List
	err := <-channels.ServiceList.Error
	nonCriticalErrors, criticalError := kdErrors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	namespaces := <-channels.NamespaceList.List
	err = <-channels.NamespaceList.Error
	nonCriticalErrors, criticalError = kdErrors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	destinationRules := <-channels.DestinationRuleList.List
	err = <-channels.DestinationRuleList.Error
	nonCriticalErrors, criticalError = kdErrors.AppendError(err, nonCriticalErrors)
	if criticalError != nil {
		return nil, criticalError
	}

	return toAppList(services.Items, destinationRules.Items, namespaces.Items, nonCriticalErrors, dsQuery)
}

// toAppList merges k8s services & istio destinationRules & virtualServices, return apps.
func toAppList(services []v1.Service, dRules []istioApi.DestinationRule, namespaces []v1.Namespace,
	nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) (*api.AppList, error) {
	var apps []*api.App

	if services == nil {
		return nil, errors.New("services not provided")
	}

	serviceList := service.CreateServiceList(services, nonCriticalErrors, dsQuery)
	for _, svc := range serviceList.Services {
		// use name+namespace to locate the app
		apps = append(apps, &api.App{
			ObjectMeta: svc.ObjectMeta,
			TypeMeta: api2.TypeMeta{
				Kind: api2.ResourceKindApp,
			},
			Destinations: []api.Destination{
				{
					Version:  "default",
					Selector: svc.Selector,
				},
			},
		})
	}

	if dRules != nil {
		for _, app := range apps {
			app.Destinations = getAppDestinations(app, dRules)
		}
	}

	// add these applications' istio statuses
	if namespaces != nil {
		for _, app := range apps {
			ns := app.ObjectMeta.Namespace
			for _, n := range namespaces {
				if n.ObjectMeta.Name == ns {
					if n.ObjectMeta.Labels["istio-injection"] == "enabled" {
						app.Status.Istio = true
					}
					break
				}
			}
		}
	}

	list := &api.AppList{
		ListMeta: api2.ListMeta{
			TotalItems: len(services),
		},
		Apps:   apps,
		Errors: nonCriticalErrors,
	}
	return list, nil
}
