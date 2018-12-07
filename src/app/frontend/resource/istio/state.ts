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

import {Ng2StateDeclaration} from '@uirouter/angular';

import {IstioComponent} from './component';

export const istioFutureState: Ng2StateDeclaration = {
  name: 'istio.**',
  url: '/istio',
  loadChildren: './resource/istio/module#IstioModule',
  data: {
    kdBreadcrumbs: {
      label: 'Istio',
    }
  },
};

export const istioState: Ng2StateDeclaration = {
  parent: 'chrome',
  name: 'istio',
  url: '/istio',
  component: IstioComponent,
  data: {
    kdBreadcrumbs: {
      label: 'Istio',
    }
  },
};
