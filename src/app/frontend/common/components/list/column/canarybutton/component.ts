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
import {Subscription} from 'rxjs';
import {VerberService} from '../../../../services/global/verber';

@Component({
  selector: 'kd-canary-button',
  templateUrl: './template.html',
})
export class CanaryButtonComponent implements ActionColumn {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;

  private onClickSubscription_: Subscription;

  constructor(private readonly verber_: VerberService, private readonly state_: StateService) {}

  setObjectMeta(objectMeta: ObjectMeta): void {
    this.objectMeta = objectMeta;
  }

  setTypeMeta(typeMeta: TypeMeta): void {
    this.typeMeta = typeMeta;
  }

  ngOnDestroy(): void {
    if (this.onClickSubscription_) this.onClickSubscription_.unsubscribe();
  }

  onClick(): void {
    this.onClickSubscription_ = this.verber_.onDeployment.subscribe(this.onSuccess_.bind(this));
    this.verber_.showDeploymentDialog(this.typeMeta.kind, this.typeMeta, this.objectMeta);
  }

  private onSuccess_(): void {
    this.state_.reload();
  }
}
