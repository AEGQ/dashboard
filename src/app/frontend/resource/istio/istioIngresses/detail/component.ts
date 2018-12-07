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

import {Component, ElementRef, OnDestroy, OnInit, QueryList, ViewChild, ViewChildren} from '@angular/core';
import {IstioIngress, Service, ServiceList} from '@api/backendapi';
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
  @ViewChild('appName') appName: ElementRef;
  @ViewChild('link1') link1: ElementRef;
  @ViewChild('protocolName') protocolName: ElementRef;
  @ViewChild('link2') link2: ElementRef;
  @ViewChild('link3', {read: ElementRef}) link3: ElementRef;
  @ViewChild('link4', {read: ElementRef}) link4: ElementRef;
  @ViewChildren('matchRule') matchRuleDoms!: QueryList<ElementRef>;

  private istioIngressDetailSubscription_: Subscription;
  private istioIngressSetName_: string;
  private readonly dynamicColumns_: ColumnWhenCondition[] = [];
  private readonly kdState_: KdStateService;
  istioIngress: IstioIngress;
  JSON: JSON;
  isInitialized = false;
  displayedColumns: string[] =
      ['name', 'labels', 'clusterip', 'internalendp', 'externalendp', 'age'];
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
            .get(EndpointManager.resource(Resource.istioIngress, true).detail(), this.istioIngressSetName_)
            .startWith({})
            .subscribe((d: IstioIngress) => {
              this.istioIngress = d;
              this.notifications_.pushErrors(d.errors);
              this.actionbar_.onInit.emit(new ResourceMeta('Istio App', d.objectMeta, d.typeMeta));
              this.isInitialized = true;
              if (Object.keys(this.istioIngress).length !== 0) {
                this.drawLineBetweenTwoElement(this.appName, this.protocolName, this.link1);
              }
              setTimeout(() => {
              }, 0);
            });
  }

  drawLineBetweenTwoElement(start: ElementRef, end: ElementRef, svgName: ElementRef): void {
    const startY = '' +
        (start.nativeElement.getBoundingClientRect().top -
         svgName.nativeElement.getBoundingClientRect().top +
         start.nativeElement.getBoundingClientRect().height / 2);

    const endX =
        (end.nativeElement.getBoundingClientRect().left -
         svgName.nativeElement.getBoundingClientRect().left);

    const endY = '' +
        (end.nativeElement.getBoundingClientRect().top -
         svgName.nativeElement.getBoundingClientRect().top +
         end.nativeElement.getBoundingClientRect().height / 2);

    const newLine = document.createElementNS('http://www.w3.org/2000/svg', 'line');
    newLine.setAttribute('id', 'line');
    newLine.setAttribute('x1', '0');
    newLine.setAttribute('y1', startY);
    newLine.setAttribute('x2', '' + (endX - 40));
    newLine.setAttribute('y2', endY);
    newLine.setAttribute('stroke', '#2b62e2');
    newLine.setAttribute('stroke-width', '4');
    newLine.setAttribute('marker-end', 'url(#arrow)');
    svgName.nativeElement.append(newLine);
  }

  map(serviceList: ServiceList): Service[] {
    return serviceList ? serviceList.services : [];
  }

  getDetailsHref(resourceName: string, namespace?: string): string {
    return this.kdState_.href('service', resourceName, namespace);
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
  onTakeOver(version: string): void {
    this.verber_.onTakeOver.subscribe((result: boolean) => {
      if (result) {
        window.location.reload();
      }
    });
    this.verber_.showTakeOverDialog(version, this.istioIngress.typeMeta, this.istioIngress.objectMeta);
  }
  onOffline(version: string): void {
    this.verber_.onOffline.subscribe((result: boolean) => {
      if (result) {
        // this.getDetail();
        window.location.reload();
      }
    });
    this.verber_.showOffLineDialog(version, this.istioIngress.typeMeta, this.istioIngress.objectMeta);
  }
  ngOnDestroy(): void {
    this.istioIngressDetailSubscription_.unsubscribe();
  }
}
