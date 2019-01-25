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

import {NgModule} from '@angular/core';

import {SharedModule} from '../../shared.module';
import {ComponentsModule} from '../components/module';

import {AlertDialog} from './alert/dialog';
import {DeleteResourceDialog} from './deleteresource/dialog';
import {DeploymentDialog} from './deployment/dialog';
import {LogsDownloadDialog} from './download/dialog';
import {EditResourceDialog} from './editresource/dialog';
import {IstioItDialog} from './istio/dialog';
import {OfflineResourceDialog} from './offlineresource/dialog';
import {RedeployResourceDialog} from './redeployresource/dialog';
import {ScaleResourceDialog} from './scaleresource/dialog';
import {TakeOverDialog} from './takeOver/dialog';

@NgModule({
  imports: [
    SharedModule,
    ComponentsModule,
  ],
  declarations: [
    AlertDialog,
    EditResourceDialog,
    DeleteResourceDialog,
    RedeployResourceDialog,
    LogsDownloadDialog,
    DeploymentDialog,
    IstioItDialog,
    TakeOverDialog,
    OfflineResourceDialog,
    ScaleResourceDialog,
  ],
  exports: [
    AlertDialog,
    EditResourceDialog,
    DeleteResourceDialog,
    RedeployResourceDialog,
    LogsDownloadDialog,
    DeploymentDialog,
    IstioItDialog,
    TakeOverDialog,
    OfflineResourceDialog,
    ScaleResourceDialog,
  ],
  entryComponents: [
    AlertDialog,
    EditResourceDialog,
    DeleteResourceDialog,
    RedeployResourceDialog,
    LogsDownloadDialog,
    DeploymentDialog,
    IstioItDialog,
    TakeOverDialog,
    OfflineResourceDialog,
    ScaleResourceDialog,
  ]
})
export class DialogsModule {
}
