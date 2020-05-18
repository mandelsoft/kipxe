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

package machines

import (
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
)

type reconciler struct {
	reconcile.DefaultReconciler

	controller controller.Interface
	config     *Config

	machines *Machines
}

var _ reconcile.Interface = &reconciler{}

func (this *reconciler) Setup() {
	this.machines.Setup(this.controller)
}

///////////////////////////////////////////////////////////////////////////////

func (this *reconciler) Config() *Config {
	return this.config
}

func (this *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	logger.Infof("reconcile")
	_, err := this.machines.Update(logger, obj)
	return reconcile.DelayOnError(logger, err)
}

func (this *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	logger.Infof("deleted")
	this.machines.Delete(logger, key.ObjectName())
	return reconcile.Succeeded(logger)
}
