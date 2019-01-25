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

import {Component, OnDestroy, OnInit} from '@angular/core';
import {IstioIngress} from '@api/backendapi';
import {ColumnWhenCondition} from '@api/frontendapi';
import {StateService} from '@uirouter/core';
import {Subscription} from 'rxjs/Subscription';

import {ActionbarService, ResourceMeta} from '../../../../common/services/global/actionbar';
import {GlobalServicesModule} from '../../../../common/services/global/module';
import {NotificationsService} from '../../../../common/services/global/notifications';
import {KdStateService} from '../../../../common/services/global/state';
import {VerberService} from '../../../../common/services/global/verber';
import {EndpointManager, Resource} from '../../../../common/services/resource/endpoint';
import {NamespacedResourceService} from '../../../../common/services/resource/resource';

@Component({
  selector: 'kd-destination-rule-detail',
  templateUrl: './template.html',
  styleUrls: ['./style.css']
})
export class IstioIngressComponent implements OnInit, OnDestroy {
  private istioIngressDetailSubscription_: Subscription;
  private istioIngressSetName_: string;
  private readonly dynamicColumns_: ColumnWhenCondition[] = [];
  private readonly kdState_: KdStateService;
  istioIngress: IstioIngress;
  JSON: JSON;
  isInitialized = false;
  columnsToDisplay = ['name', 'namespace', 'action'];
  constructor(
      private readonly istioIngress_: NamespacedResourceService<IstioIngress>,
      private readonly actionbar_: ActionbarService,
      private readonly state_: StateService,
      private readonly notifications_: NotificationsService,
      private readonly verber_: VerberService,
  ) {
    this.JSON = JSON;
    this.kdState_ = GlobalServicesModule.injector.get(KdStateService);
  }

  ngOnInit(): void {
    this.getDetail();
  }

  getDetail(): void {
    this.istioIngressSetName_ = this.state_.params.resourceName;
    this.istioIngressDetailSubscription_ =
        this.istioIngress_
            .get(
                EndpointManager.resource(Resource.istioIngress, true).detail(),
                this.istioIngressSetName_)
            .startWith({})
            .subscribe((d: IstioIngress) => {
              this.istioIngress = d;
              this.notifications_.pushErrors(d.errors);
              this.actionbar_.onInit.emit(
                  new ResourceMeta('Istio Ingress', d.objectMeta, d.typeMeta));
              this.isInitialized = true;
            });
  }

  shouldShowColumn(dynamicColName: string): boolean {
    const col = this.dynamicColumns_.find((condition) => {
      return condition.col === dynamicColName;
    });
    if (col !== undefined) {
      return col.whenCallback();
    }

    return false;
  }

  ngOnDestroy(): void {
    this.istioIngressDetailSubscription_.unsubscribe();
  }
}
