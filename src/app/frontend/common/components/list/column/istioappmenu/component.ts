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

import {Component, Input} from '@angular/core';
import {ObjectMeta, TypeMeta} from '@api/backendapi';
import {ActionColumn} from '@api/frontendapi';
import {StateService} from '@uirouter/core';
import {Subscription} from "rxjs";
import {logsState} from '../../../../../logs/state';
import {LogsStateParams} from '../../../../params/params';
import {VerberService} from "../../../../services/global/verber";

@Component({
  selector: 'kd-istio-app-menu',
  templateUrl: './template.html',
})
export class IstioAppMenuComponent implements ActionColumn {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;

  private onDeleteSubscription_: Subscription;

  constructor(private readonly verber_: VerberService, private readonly state_: StateService) {}

  setObjectMeta(objectMeta: ObjectMeta): void {
    this.objectMeta = objectMeta;
  }

  setTypeMeta(typeMeta: TypeMeta): void {
    this.typeMeta = typeMeta;
  }

  onDelete(): void {
    this.onDeleteSubscription_ = this.verber_.onIstioDelete.subscribe(this.onSuccess_.bind(this));
    this.verber_.showIstioDeleteDialog(this.typeMeta, this.objectMeta);
  }

  private onSuccess_(): void {
    this.state_.reload();
  }
}
