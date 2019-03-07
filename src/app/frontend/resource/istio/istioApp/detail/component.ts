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

import {HttpClient} from '@angular/common/http';
import {Component, ElementRef, OnDestroy, OnInit, QueryList, Renderer2, ViewChild, ViewChildren} from '@angular/core';
import {DomSanitizer, SafeUrl} from '@angular/platform-browser';
import {IstioApp, RawDeployment, Service, ServiceList} from '@api/backendapi';
import {ColumnWhenCondition} from '@api/frontendapi';
import {StateService} from '@uirouter/core';
import {Observable} from 'rxjs/Rx';
import {Subscription} from 'rxjs/Subscription';
import * as util from 'util';

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
export class IstioAppComponent implements OnInit, OnDestroy {
  @ViewChild('appName') appName: ElementRef;
  @ViewChild('link1') link1: ElementRef;
  @ViewChild('link2') link2: ElementRef;
  @ViewChild('link3', {read: ElementRef}) link3: ElementRef;
  @ViewChild('http') http: ElementRef;
  @ViewChild('tcp') tcp: ElementRef;
  @ViewChild('tls') tls: ElementRef;
  @ViewChildren('matchRule') matchRuleDoms!: QueryList<ElementRef>;

  private istioAppDetailSubscription_: Subscription;
  private pollingData: Subscription;
  private istioAppSetName_: string;
  private readonly dynamicColumns_: ColumnWhenCondition[] = [];
  private readonly kdState_: KdStateService;

  istioApp: IstioApp;
  metrics: SafeUrl[] = [];
  columnsToDisplay = ['name', 'namespace', 'action'];

  JSON: JSON;
  isInitialized = false;
  refreshed = false;

  constructor(
      private readonly istioApp_: NamespacedResourceService<IstioApp>,
      private readonly actionbar_: ActionbarService, private readonly state_: StateService,
      private readonly notifications_: NotificationsService,
      private readonly verber_: VerberService, readonly sanitizer: DomSanitizer,
      private readonly httpClient: HttpClient, private readonly renderer: Renderer2) {
    this.JSON = JSON;
    this.kdState_ = GlobalServicesModule.injector.get(KdStateService);
    this.httpClient = httpClient;
    this.renderer = renderer;
  }

  ngOnInit(): void {
    this.getDetail();
  }

  getDetail(): void {
    this.istioAppSetName_ = this.state_.params.resourceName;
    this.istioAppDetailSubscription_ =
        this.istioApp_
            .get(EndpointManager.resource(Resource.istioApp, true).detail(), this.istioAppSetName_)
            .startWith({})
            .subscribe((d: IstioApp) => {
              this.istioApp = d;
              if (d.metrics) {
                this.metrics = [
                  this.sanitizer.bypassSecurityTrustResourceUrl(d.metrics.clientQps),
                  this.sanitizer.bypassSecurityTrustResourceUrl(d.metrics.clientLatency),
                  this.sanitizer.bypassSecurityTrustResourceUrl(d.metrics.serverQps),
                  this.sanitizer.bypassSecurityTrustResourceUrl(d.metrics.serverLatency)
                ];
              }
              this.notifications_.pushErrors(d.errors);
              this.actionbar_.onInit.emit(new ResourceMeta('Istio App', d.objectMeta, d.typeMeta));
              this.isInitialized = true;
              setTimeout(() => {
                this.drawLines();
                this.refreshState();
              }, 0);
            });
  }

  // refreshState polling the application's deployments every 5 seconds and change the destinations'
  // color.
  refreshState(): void {
    if (!this.istioApp.objectMeta || this.refreshed) {
      return;
    }
    this.pollingData =
        Observable.interval(5000)
            .startWith(100)
            .switchMap(
                () => this.httpClient.get(
                    '/api/v1/istio/app/' + this.istioApp.objectMeta.namespace + '/' +
                    this.istioApp.objectMeta.name + '/deployments'))
            .subscribe((data: []) => {
              data.forEach((dep: RawDeployment) => {
                let subset = '';
                if (dep.metadata.labels.version) {
                  subset = dep.metadata.labels.version;
                }

                const cls = 'success-destination';
                const element = new ElementRef(document.getElementById('app-destination-' + subset))
                                    .nativeElement;
                if (this.deploymentStatus(dep)) {
                  this.renderer.addClass(element, cls);
                } else {
                  this.renderer.removeClass(element, cls);
                }
              });
            });
    this.refreshed = true;
  }

  deploymentStatus(dep: RawDeployment): boolean {
    if (dep.metadata.generation <= dep.status.observedGeneration) {
      if (dep.spec.replicas != null && dep.status.updatedReplicas < dep.spec.replicas) {
        console.log('wait for deployment rollout to finish');
        return false;
      }

      if (dep.status.replicas > dep.status.updatedReplicas) {
        console.log(
            'Waiting for deployment rollout to finish: old replicas are pending termination...');
        return false;
      }

      if (!dep.status.availableReplicas ||
          dep.status.availableReplicas < dep.status.updatedReplicas) {
        console.log('Waiting for deployment rollout to finish: updated replicas are available...');
        return false;
      }
      return true;
    }
    return false;
  }

  drawLines(): void {
    const lines: ElementRef[][] = [];

    if (this.istioApp.virtualServices) {
      this.istioApp.virtualServices.forEach((v, i) => {
        const sourceNodes: ElementRef[] = [];
        if (v.hosts) {
          v.hosts.forEach((_, j) => {
            sourceNodes.push(new ElementRef(document.getElementById('app-host-' + i + '-' + j)));
          });
        }

        if (v.gateways) {
          v.gateways.forEach((_, j) => {
            sourceNodes.push(new ElementRef(document.getElementById('app-gateway-' + i + '-' + j)));
          });
        }

        if (v.http) {
          v.http.forEach((h, j) => {
            const matchNode: ElementRef =
                new ElementRef(document.getElementById('app-operator-' + i + '-' + j));
            h.route.forEach((r) => {
              if (this.FQDN(r.destination.host, v.objectMeta.namespace) !==
                  this.FQDN(this.istioApp.objectMeta.name, this.istioApp.objectMeta.namespace)) {
                return;
              }

              const subset = r.destination.subset;
              let destinationNode: ElementRef;
              if (subset) {
                destinationNode =
                    new ElementRef(document.getElementById('app-destination-' + subset));
              } else {
                destinationNode = new ElementRef(document.getElementById('app-any-destination'));
              }
              lines.push([matchNode, destinationNode, this.link3]);
            });

            lines.push([this.http, matchNode, this.link2]);
          });

          Array.from(sourceNodes).forEach((sn: ElementRef, _) => {
            lines.push([sn, this.http, this.link1]);
          });
        }
      });
    }

    lines.forEach((e, _) => {
      this.drawLineBetweenTwoElement(e[0], e[1], e[2]);
    });
  }

  FQDN(name: string, namespace: string): string {
    if (name.endsWith('.svc.cluster.local')) {
      return name;
    }
    return util.format('%s.%s.svc.cluster.local', name, namespace);
  }

  drawLineBetweenTwoElement(start: ElementRef, end: ElementRef, svgName: ElementRef): void {
    const startY =
        (start.nativeElement.getBoundingClientRect().top -
         svgName.nativeElement.getBoundingClientRect().top +
         start.nativeElement.getBoundingClientRect().height / 2);

    const endX =
        (end.nativeElement.getBoundingClientRect().left -
         svgName.nativeElement.getBoundingClientRect().left);

    const endY =
        (end.nativeElement.getBoundingClientRect().top -
         svgName.nativeElement.getBoundingClientRect().top +
         end.nativeElement.getBoundingClientRect().height / 2);

    const newLine = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    const d = `M${0} ${startY} C ${endX / 2} ${startY} ${endX / 2} ${endY} ${endX - 10} ${endY}`;
    newLine.setAttribute('d', d);
    newLine.setAttribute('stroke-width', '2');
    newLine.setAttribute('stroke', '#9fc1e2');
    newLine.setAttribute('fill', 'transparent');
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
    this.verber_.showTakeOverDialog(version, this.istioApp.typeMeta, this.istioApp.objectMeta);
  }

  onOffline(version: string): void {
    this.verber_.onOffline.subscribe((result: boolean) => {
      if (result) {
        // this.getDetail();
        window.location.reload();
      }
    });
    this.verber_.showOffLineDialog(version, this.istioApp.typeMeta, this.istioApp.objectMeta);
  }

  ngOnDestroy(): void {
    this.istioAppDetailSubscription_.unsubscribe();
    this.pollingData.unsubscribe();
  }
}
