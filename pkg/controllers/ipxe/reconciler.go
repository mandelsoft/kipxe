/*
 * Copyright 2020 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package ipxe

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
)

type reconciler struct {
	reconcile.DefaultReconciler

	controller controller.Interface
	config     *Config
	infobase   *InfoBase
}

var _ reconcile.Interface = &reconciler{}

func (this *reconciler) Setup() {
	this.infobase.Setup()
}

///////////////////////////////////////////////////////////////////////////////

func (this *reconciler) Config(cfg interface{}) *Config {
	return this.config
}

func (this *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	var err error
	logger.Infof("reconcile")
	switch obj.Data().(type) {
	case *v1alpha1.Profile:
		_, err = this.infobase.profiles.Update(logger, obj)
	case *v1alpha1.Matcher:
		_, err = this.infobase.matchers.Update(logger, obj)
	case *v1alpha1.Document:
		_, err = this.infobase.documents.Update(logger, obj)
	}
	return reconcile.DelayOnError(logger, err)
}

func (this *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	logger.Infof("deleted")
	switch key.GroupKind() {
	case v1alpha1.PROFILE:
		this.infobase.profiles.Delete(logger, key.ObjectName())
	case v1alpha1.MATCHER:
		this.infobase.matchers.Delete(logger, key.ObjectName())
	case v1alpha1.DOCUMENT:
		this.infobase.documents.Delete(logger, key.ObjectName())
	}
	return reconcile.Succeeded(logger)
}
