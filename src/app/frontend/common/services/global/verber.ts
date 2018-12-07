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

import {HttpClient, HttpErrorResponse, HttpHeaders} from '@angular/common/http';
import {EventEmitter, Inject, Injectable} from '@angular/core';
import {MatDialog, MatDialogConfig} from '@angular/material';
import {ObjectMeta, TypeMeta} from '@api/backendapi';

import {Config, CONFIG_DI_TOKEN} from '../../../index.config';
import {AlertDialog, AlertDialogConfig} from '../../dialogs/alert/dialog';
import {DeleteResourceDialog} from '../../dialogs/deleteresource/dialog';
import {DeploymentDialog} from '../../dialogs/deployment/dialog';
import {EditResourceDialog} from '../../dialogs/editresource/dialog';
import {IstioItDialog} from '../../dialogs/istio/dialog';
import {OfflineResourceDialog} from '../../dialogs/offlineresource/dialog';
import {TakeOverDialog} from '../../dialogs/takeOver/dialog';
import {IstioResource} from "../../resources/istioresource";
import {RawResource} from '../../resources/rawresource';
import {ResourceMeta} from './actionbar';
import {CsrfTokenService} from './csrftoken';

@Injectable()
export class VerberService {
  onDelete = new EventEmitter<boolean>();
  onIstioDelete = new EventEmitter<boolean>();
  onEdit = new EventEmitter<boolean>();
  onDeployment = new EventEmitter<boolean>();
  onOffline = new EventEmitter<boolean>();
  onTakeOver = new EventEmitter<boolean>();
  onIstioIt = new EventEmitter<boolean>();

  constructor(
      private readonly dialog_: MatDialog,
      private readonly http_: HttpClient,
      private readonly csrfToken_: CsrfTokenService,
      @Inject(CONFIG_DI_TOKEN) private readonly CONFIG: Config,
  ) {}

  showDeleteDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_.open(DeleteResourceDialog, dialogConfig).afterClosed().subscribe((doDelete) => {
      if (doDelete) {
        const url = RawResource.getUrl(typeMeta, objectMeta);
        this.http_.delete(url).subscribe(() => {
          this.onDelete.emit(true);
        }, this.handleErrorResponse_.bind(this));
      }
    });
  }

  showOffLineDialog(version: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(version, typeMeta, objectMeta);
    this.dialog_.open(OfflineResourceDialog, dialogConfig)
        .afterClosed()
        .subscribe(async (doOffline) => {
          if (doOffline) {
            const url = IstioResource.getUrl(typeMeta, objectMeta);
            const {token} = await this.csrfToken_.getTokenForAction('istio').toPromise();
            this.http_.delete(url+`/${version}`, {headers: {[this.CONFIG.csrfHeaderName]: token}})
                .subscribe(() => {
                  this.onOffline.emit(true);
                }, this.handleErrorResponse_.bind(this));
          }
        });
  }

  showIstioDeleteDialog(typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_("", typeMeta, objectMeta);
    this.dialog_.open(DeleteResourceDialog, dialogConfig).afterClosed().subscribe(async (doDelete) => {
      if (doDelete) {
        const url = IstioResource.getUrl(typeMeta, objectMeta);
        this.http_.delete(url).subscribe(() => {
          this.onIstioDelete.emit(true);
        }, this.handleErrorResponse_.bind(this));
      }
    });
  }

  showTakeOverDialog(version: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(version, typeMeta, objectMeta);
    this.dialog_.open(TakeOverDialog, dialogConfig).afterClosed().subscribe(async (doOffline) => {
      if (doOffline) {
        const url = IstioResource.getUrl(typeMeta, objectMeta)+`/${version}/takeover`;
        const {token} = await this.csrfToken_.getTokenForAction('istio').toPromise();
        this.http_.post(url, {}, {headers: {[this.CONFIG.csrfHeaderName]: token}}).subscribe(() => {
          this.onTakeOver.emit(true);
        }, this.handleErrorResponse_.bind(this));
      }
    });
  }

  showEditDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_.open(EditResourceDialog, dialogConfig).afterClosed().subscribe((result) => {
      if (result) {
        const url = RawResource.getUrl(typeMeta, objectMeta);
        this.http_.put(url, JSON.parse(result), {headers: this.getHttpHeaders_()}).subscribe(() => {
          this.onEdit.emit(true);
        }, this.handleErrorResponse_.bind(this));
      }
    });
  }

  showDeploymentDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_.open(DeploymentDialog, dialogConfig).afterClosed().subscribe(async (result) => {
      if (result) {
        const {token} = await this.csrfToken_.getTokenForAction('istio').toPromise();
        const url = IstioResource.getUrl(typeMeta, objectMeta)+`/canary`;
        result = JSON.parse(result);
        result.podTemplate = JSON.parse(result.podTemplateJSON);
        this.http_.post(url, result, {headers: {[this.CONFIG.csrfHeaderName]: token}})
            .subscribe(() => {
              this.onDeployment.emit(true);
            }, this.handleErrorResponse_.bind(this));
      }
    });
  }

  showIstioItDialog(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta): void {
    const dialogConfig = this.getDialogConfig_(displayName, typeMeta, objectMeta);
    this.dialog_.open(IstioItDialog, dialogConfig).afterClosed().subscribe(async (result) => {
      if (result) {
        const {token} = await this.csrfToken_.getTokenForAction('istio').toPromise();
        const url = IstioResource.getUrl(typeMeta, objectMeta)+`/istio-it`;
        this.http_.post(url, result, {headers: {[this.CONFIG.csrfHeaderName]: token}})
            .subscribe(() => {
              this.onDeployment.emit(true);
            }, this.handleErrorResponse_.bind(this));
      }
    });
  }

  getDialogConfig_(displayName: string, typeMeta: TypeMeta, objectMeta: ObjectMeta):
      MatDialogConfig<ResourceMeta> {
    return {width: '630px', data: {displayName, typeMeta, objectMeta}};
  }

  handleErrorResponse_(err: HttpErrorResponse): void {
    if (err) {
      const alertDialogConfig: MatDialogConfig<AlertDialogConfig> = {
        width: '630px',
        data: {
          title: err.statusText === 'OK' ? 'Internal server error' : err.statusText,
          // TODO Add || this.localizerService_.localize(err.data).
          message: err.error || 'Could not perform the operation.',
          confirmLabel: 'OK',
        }
      };
      this.dialog_.open(AlertDialog, alertDialogConfig);
    }
  }

  getHttpHeaders_(): HttpHeaders {
    const headers = new HttpHeaders();
    headers.set('Content-Type', 'application/json');
    headers.set('Accept', 'application/json');
    return headers;
  }
}
