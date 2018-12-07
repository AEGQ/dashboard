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

import {HttpParams} from '@angular/common/http';
import {Component, ComponentFactoryResolver, Input} from '@angular/core';
import {IstioApp, IstioAppList} from '@api/backendapi';
import {StateService} from '@uirouter/core';
import {Observable} from 'rxjs/Observable';

import {istioAppState} from '../../../../resource/istio/istioApp/state';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NamespaceService} from '../../../services/global/namespace';
import {NotificationsService} from '../../../services/global/notifications';
import {VerberService} from '../../../services/global/verber';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {CanaryButtonComponent} from "../../list/column/canarybutton/component";
import {IstioAppMenuComponent} from "../../list/column/istioappmenu/component";
import {LogsButtonComponent} from "../../list/column/logsbutton/component";
import {ListGroupIdentifiers, ListIdentifiers} from '../groupids';

@Component({selector: 'kd-istio-app-list', templateUrl: './template.html'})
export class IstioAppListComponent extends ResourceListWithStatuses<IstioAppList, IstioApp> {
  @Input() endpoint = EndpointManager.resource(Resource.istioApp, true).list();
  constructor(
      state: StateService,
      private readonly istioApp_: NamespacedResourceService<IstioAppList>,
      resolver: ComponentFactoryResolver,
      notifications: NotificationsService,
      private readonly namespaceService_: NamespaceService,
      private readonly verber_: VerberService,
  ) {
    super(istioAppState.name, state, notifications, resolver);
    this.id = ListIdentifiers.statefulSet;
    this.groupId = ListGroupIdentifiers.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<CanaryButtonComponent>('canary', CanaryButtonComponent);
    this.registerActionColumn<IstioAppMenuComponent>('menu', IstioAppMenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<IstioAppList> {
    return this.istioApp_.get(this.endpoint, undefined, params);
  }

  map(istioAppList: IstioAppList): IstioApp[] {
    istioAppList.apps.forEach(app => {
      app.destinationVersions = app.destinations.map(destination => {
        return destination.version;
      });
    });
    return istioAppList.apps;
  }

  isInErrorState(resource: IstioApp): boolean {
    return !!resource.errors && resource.errors.length > 0;
  }

  isInPendingState(resource: IstioApp): boolean {
    return !!resource.errors && resource.errors.length === 0;
  }

  isInSuccessState(resource: IstioApp): boolean {
    return !resource.errors || resource.errors.length === 0;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'versions', 'age'];
  }

  hasErrors(statefulSet: IstioApp): boolean {
    return !!statefulSet.errors && statefulSet.errors.length > 0;
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
