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
import {Component, Inject, OnInit, ViewChild} from '@angular/core';
import {MAT_DIALOG_DATA, MatButtonToggleGroup, MatDialogRef} from '@angular/material';
import {CanaryDeployment, CanaryDeploymentInput} from '@api/backendapi';
import {dump as toYaml, load as fromYaml} from 'js-yaml';

import {ResourceMeta} from '../../services/global/actionbar';

enum EditorMode {
  JSON = 'json',
  YAML = 'yaml',
}

@Component({
  selector: 'kd-deployment-dialog',
  templateUrl: 'template.html',
  styles: [
    '.deployment-form { min-width: 150px; max-width: 500px; width: 100%; } .deployment-full-width { width: 100%; }'
  ]
})
export class DeploymentDialog implements OnInit {
  private selectedMode_ = EditorMode.YAML;
  @ViewChild('group') buttonToggleGroup: MatButtonToggleGroup;
  text = '';
  modes = EditorMode;
  development: CanaryDeploymentInput =
      {version: '', replicas: null, description: '', podTemplate: '', podTemplateJSON: ''};

  constructor(
      public dialogRef: MatDialogRef<DeploymentDialog>,
      @Inject(MAT_DIALOG_DATA) public data: ResourceMeta, private readonly http_: HttpClient) {}

  ngOnInit(): void {
    const url = `api/v1/istio/${this.data.typeMeta.kind}/${this.data.objectMeta.namespace}/${
        this.data.objectMeta.name}/deployments`;
    this.http_.get(url).toPromise().then((response: CanaryDeployment[]) => {
      this.development.replicas = response[0].spec.replicas;
      this.development.podTemplate = this.toRawJSON(response[0].spec.template);
      this.development.version = '';
      this.development.description = '';
      this.development.podTemplateJSON = this.toRawJSON(response[0].spec.template);
    });

    this.buttonToggleGroup.valueChange.subscribe((selectedMode: EditorMode) => {
      this.selectedMode_ = selectedMode;

      if (this.development.podTemplate) {
        this.updateText();
      }
    });
  }

  onNoClick(): void {
    this.dialogRef.close();
  }

  getJSON(): string {
    if (this.selectedMode_ === EditorMode.YAML) {
      this.development.podTemplateJSON = this.toRawJSON(fromYaml(this.development.podTemplate));
    } else {
      this.development.podTemplateJSON = this.development.podTemplate;
    }
    return this.toRawJSON(this.development);
  }

  getSelectedMode(): string {
    return this.buttonToggleGroup.value;
  }

  private updateText(): void {
    if (this.selectedMode_ === EditorMode.YAML) {
      this.development.podTemplate = toYaml(JSON.parse(this.development.podTemplate));
    } else {
      this.development.podTemplate = this.toRawJSON(fromYaml(this.development.podTemplate));
    }
  }

  private toRawJSON(object: {}): string {
    return JSON.stringify(object, null, '\t');
  }
}
