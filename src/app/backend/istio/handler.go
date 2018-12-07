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

package istio

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	clientapi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	kdErrors "github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/istio/api"
	"github.com/kubernetes/dashboard/src/app/backend/istio/app"
	"github.com/kubernetes/dashboard/src/app/backend/istio/ingress"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/virtualservice"
	"k8s.io/api/apps/v1beta1"
)

// IstioHandler manages all endpoints related to istio management.
type IstioHandler struct {
	cManager clientapi.ClientManager
}

// Install creates new endpoints for istio management.
func (self *IstioHandler) Install(ws *restful.WebService) {
	ws.Route(
		ws.GET("/istio/app").
			To(self.handleGetApps).
			Writes(api.AppList{}))
	ws.Route(
		ws.GET("/istio/app/{namespace}").
			To(self.handleGetApps).
			Writes(api.AppList{}))
	ws.Route(
		ws.GET("/istio/app/{namespace}/{app}").
			To(self.handleGetAppDetail).
			Writes(api.App{}))
	ws.Route(
		ws.POST("/istio/app/{namespace}/{app}").
			To(self.handleCreateApp).
			Writes(nil))
	ws.Route(
		ws.DELETE("/istio/app/{namespace}/{app}").
			To(self.handleDeleteApp).
			Writes(nil))
	ws.Route(
		ws.POST("/istio/app/{namespace}/{app}/canary").
			To(self.handleCanaryApp).
			Writes(nil))

	ws.Route(
		ws.DELETE("/istio/app/{namespace}/{app}/{version}").
			To(self.handleOfflineVersion).
			Writes(nil))
	ws.Route(
		ws.POST("/istio/app/{namespace}/{app}/{version}/takeover").
			To(self.handleAppTakeOverAllTraffic).
			Writes(nil))
	ws.Route(
		ws.GET("/istio/app/{namespace}/{app}/versions").
			To(self.handleGetAppVersions).
			Writes(nil))
	ws.Route(
		ws.GET("/istio/app/{namespace}/{app}/deployments").
			To(self.handleGetAppDeploySpec).
			Writes([]v1beta1.Deployment{}))

	// Istio Ingresses
	ws.Route(
		ws.GET("/istio/ingress/{namespace}").
			To(self.handleGetIngresses).
			Writes(nil))
	ws.Route(
		ws.POST("/istio/ingress/{namespace}/{name}").
			To(self.handleCreateIngress).
			Writes(nil))
	ws.Route(
		ws.GET("/istio/ingress/{namespace}/{name}").
			To(self.handleGetIngress).
			Writes(nil))
	ws.Route(
		ws.DELETE("/istio/ingress/{namespace}/{name}").
			To(self.handleDeleteIngress).
			Writes(nil))
}

// handleGetAppVersions gets the application's versions and their flow control.
func (self *IstioHandler) handleGetAppVersions(request *restful.Request, response *restful.Response) {
}

func (self *IstioHandler) handleGetApps(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	// 1. Get Services by the namespace
	// 2. Assemble APPs
	// if service has destination rules with the same host name with service, treat them as the
	// gray release
	// app_name, versions, label, created_at
	result, err := app.GetApps(client, istioClient, parseNamespacePathParameter(request), parseDataSelectPathParameter(request))
	if err != nil {
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (self *IstioHandler) handleGetAppDetail(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	// 1. Get Services by the namespace
	// 2. Assemble APPs
	// if service has destination rules with the same host name with service, treat them as the
	// gray release
	// app_name, versions, label, created_at
	appName := request.PathParameter("app")
	dataSelect := parseDataSelectPathParameter(request)
	dataSelect.FilterQuery = dataselect.NewFilterQuery([]string{"name", appName})
	result, err := app.GetAppDetail(client, istioClient, parseNamespacePathParameter(request), appName, dataSelect)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// handleOfflineVersion makes one version of app offline
func (self *IstioHandler) handleOfflineVersion(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	appName := request.PathParameter("app")
	version := request.PathParameter("version")
	offlineType := request.QueryParameter("offlineType")
	if offlineType == "" {
		offlineType = virtualservice.OnlyHost
	}

	if err := app.OfflineAppVersion(client, istioClient, namespace, appName, version, offlineType); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, nil)
}

// handleAppTakeOverAllTraffic makes one version of app the only online version.
func (self *IstioHandler) handleAppTakeOverAllTraffic(request *restful.Request, response *restful.Response) {
	// change virtual service
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	appName := request.PathParameter("app")
	version := request.PathParameter("version")
	offlineType := request.QueryParameter("offlineType")
	if offlineType == "" {
		offlineType = virtualservice.OnlyHost
	}

	err = app.TakeOverAllTraffic(client, istioClient, namespace, appName, version, offlineType)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, nil)
}

// handleGetAppFlowControl queries the specified application's k8s virtualservice details.
func (self *IstioHandler) handleGetAppFlowControl(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	appName := request.PathParameter("app")
	namespace := parseNamespacePathParameter(request)

	deployments, err := app.GetAppDeploySpec(client, namespace, appName)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, deployments)
}

// handleGetAppDeploySpec queries the specified application's k8s deployment details.
func (self *IstioHandler) handleGetAppDeploySpec(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	appName := request.PathParameter("app")
	namespace := parseNamespacePathParameter(request)

	deployments, err := app.GetAppDeploySpec(client, namespace, appName)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, deployments)
}

func (self *IstioHandler) handleAppStatus(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusCreated, api.Status{})
}

// handleDeleteApp deletes istio application.
func (self *IstioHandler) handleDeleteApp(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	appName := request.PathParameter("app")
	namespace := parseNamespacePathParameter(request)

	if err := app.DeleteApp(client, istioClient, namespace, appName); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, nil)
}

// handleCreateApp creates istio application.
func (self *IstioHandler) handleCreateApp(request *restful.Request, response *restful.Response) {
	newApp := new(api.NewApplication)
	if err := request.ReadEntity(newApp); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	appName := request.PathParameter("app")
	namespace := parseNamespacePathParameter(request)

	if err := app.CreateApp(client, istioClient, namespace, appName, newApp); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, nil)
}

// handleCanaryApp creates application with specified version
func (self *IstioHandler) handleCanaryApp(request *restful.Request, response *restful.Response) {
	canaryDep := new(api.CanaryDeployment)
	if err := request.ReadEntity(canaryDep); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	appName := request.PathParameter("app")
	namespace := parseNamespacePathParameter(request)

	if err := app.CanaryApp(client, istioClient, namespace, appName, canaryDep); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, nil)
}

// handleGetIngresses lists the istio ingresses.
func (self *IstioHandler) handleGetIngresses(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	ingresses, err := ingress.GetIngresses(client, istioClient, namespace, parseDataSelectPathParameter(request))
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, ingresses)
}

// handleGetIngress get the ingress detail.
func (self *IstioHandler) handleGetIngress(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	name := request.PathParameter("name")
	ing, err := ingress.GetIngress(client, istioClient, namespace.ToRequestParam(), name)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, ing)
}

// handleCreateIngress create istio ingress.
func (self *IstioHandler) handleCreateIngress(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	name := request.PathParameter("name")
	if err := ingress.CreateIngress(client, istioClient, namespace, name); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusCreated, nil)
}

// handleDeleteIngress deletes istio ingress.
func (self *IstioHandler) handleDeleteIngress(request *restful.Request, response *restful.Response) {
	client, err := self.cManager.Client(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	istioClient, err := self.cManager.IstioClient(request)
	if err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	name := request.PathParameter("name")
	if err := ingress.DeleteIngress(client, istioClient, namespace.ToRequestParam(), name); err != nil {
		kdErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, nil)
}

// Parses query parameters of the request and returns a DataSelectQuery object
func parseDataSelectPathParameter(request *restful.Request) *dataselect.DataSelectQuery {
	paginationQuery := parsePaginationPathParameter(request)
	sortQuery := parseSortPathParameter(request)
	filterQuery := parseFilterPathParameter(request)
	metricQuery := parseMetricPathParameter(request)
	return dataselect.NewDataSelectQuery(paginationQuery, sortQuery, filterQuery, metricQuery)
}

func parsePaginationPathParameter(request *restful.Request) *dataselect.PaginationQuery {
	itemsPerPage, err := strconv.ParseInt(request.QueryParameter("itemsPerPage"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	page, err := strconv.ParseInt(request.QueryParameter("page"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	// Frontend pages start from 1 and backend starts from 0
	return dataselect.NewPaginationQuery(int(itemsPerPage), int(page-1))
}

func parseFilterPathParameter(request *restful.Request) *dataselect.FilterQuery {
	return dataselect.NewFilterQuery(strings.Split(request.QueryParameter("filterBy"), ","))
}

// Parses query parameters of the request and returns a SortQuery object
func parseSortPathParameter(request *restful.Request) *dataselect.SortQuery {
	return dataselect.NewSortQuery(strings.Split(request.QueryParameter("sortBy"), ","))
}

// Parses query parameters of the request and returns a MetricQuery object
func parseMetricPathParameter(request *restful.Request) *dataselect.MetricQuery {
	metricNamesParam := request.QueryParameter("metricNames")
	var metricNames []string
	if metricNamesParam != "" {
		metricNames = strings.Split(metricNamesParam, ",")
	} else {
		metricNames = nil
	}
	aggregationsParam := request.QueryParameter("aggregations")
	var rawAggregations []string
	if aggregationsParam != "" {
		rawAggregations = strings.Split(aggregationsParam, ",")
	} else {
		rawAggregations = nil
	}
	aggregationModes := metricapi.AggregationModes{}
	for _, e := range rawAggregations {
		aggregationModes = append(aggregationModes, metricapi.AggregationMode(e))
	}
	return dataselect.NewMetricQuery(metricNames, aggregationModes)

}

// parseNamespacePathParameter parses namespace selector for list pages in path parameter.
// The namespace selector is a comma separated list of namespaces that are trimmed.
// No namespaces means "view all user namespaces", i.e., everything except kube-system.
func parseNamespacePathParameter(request *restful.Request) *common.NamespaceQuery {
	namespace := request.PathParameter("namespace")
	namespaces := strings.Split(namespace, ",")
	var nonEmptyNamespaces []string
	for _, n := range namespaces {
		n = strings.Trim(n, " ")
		if len(n) > 0 {
			nonEmptyNamespaces = append(nonEmptyNamespaces, n)
		}
	}
	return common.NewNamespaceQuery(nonEmptyNamespaces)
}

// NewIstioHandler creates IstioHandler.
func NewIstioHandler(cManager clientapi.ClientManager) IstioHandler {
	return IstioHandler{cManager: cManager}
}
