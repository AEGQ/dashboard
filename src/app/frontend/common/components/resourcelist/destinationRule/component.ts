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
import {DestinationRuleList, DestinationRules} from '@api/backendapi';
import {StateService} from '@uirouter/core';
import {Observable} from 'rxjs/Observable';
import {destinationRuleState} from '../../../../resource/discovery/destinationRule/state';
import {ResourceListWithStatuses} from '../../../resources/list';
import {NamespaceService} from '../../../services/global/namespace';
import {NotificationsService} from '../../../services/global/notifications';
import {EndpointManager, Resource} from '../../../services/resource/endpoint';
import {NamespacedResourceService} from '../../../services/resource/resource';
import {MenuComponent} from '../../list/column/menu/component';
import {ListGroupIdentifiers, ListIdentifiers} from '../groupids';

@Component({selector: 'kd-destination-rule-list', templateUrl: './template.html'})
export class DestinationRuleListComponent extends
    ResourceListWithStatuses<DestinationRuleList, DestinationRules> {
  @Input() endpoint = EndpointManager.resource(Resource.destinationRule, true).list();
  constructor(
      state: StateService,
      private readonly destinationRule_: NamespacedResourceService<DestinationRuleList>,
      resolver: ComponentFactoryResolver, notifications: NotificationsService,
      private readonly namespaceService_: NamespaceService) {
    super(destinationRuleState.name, state, notifications, resolver);
    this.id = ListIdentifiers.statefulSet;
    this.groupId = ListGroupIdentifiers.workloads;

    // Register status icon handlers
    this.registerBinding(this.icon.checkCircle, 'kd-success', this.isInSuccessState);
    this.registerBinding(this.icon.timelapse, 'kd-muted', this.isInPendingState);
    this.registerBinding(this.icon.error, 'kd-error', this.isInErrorState);

    // Register action columns.
    this.registerActionColumn<MenuComponent>('menu', MenuComponent);

    // Register dynamic columns.
    this.registerDynamicColumn('namespace', 'name', this.shouldShowNamespaceColumn_.bind(this));
  }

  getResourceObservable(params?: HttpParams): Observable<DestinationRuleList> {
    return this.destinationRule_.get(this.endpoint, undefined, params);
  }

  map(destinationRuleList: DestinationRuleList): DestinationRules[] {
    return destinationRuleList.destinationRules;
  }

  isInErrorState(resource: DestinationRules): boolean {
    return !!resource.errors && resource.errors.length > 0;
  }

  isInPendingState(resource: DestinationRules): boolean {
    return !!resource.errors && resource.errors.length === 0;
  }

  isInSuccessState(resource: DestinationRules): boolean {
    return !resource.errors || resource.errors.length === 0;
  }

  getDisplayColumns(): string[] {
    return ['statusicon', 'name', 'labels', 'hosts', 'age'];
  }

  hasErrors(statefulSet: DestinationRules): boolean {
    return !!statefulSet.errors && statefulSet.errors.length > 0;
  }

  private shouldShowNamespaceColumn_(): boolean {
    return this.namespaceService_.areMultipleNamespacesSelected();
  }
}
