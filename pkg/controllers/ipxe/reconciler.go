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
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/server"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/controllers"
	"github.com/mandelsoft/kipxe/pkg/controllers/ipxe/ready"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type Ready struct{}

func (this Ready) IsReady() bool { return true }

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

func (this *reconciler) Start() {
	logger := this.controller.NewContext("server", "kipxe")
	ipxe := server.NewHTTPServer(this.controller.GetContext(), logger, "kipxe")

	infobase := &kipxe.InfoBase{
		Registry:  controllers.GetSharedRegistry(this.controller),
		Resources: this.infobase.resources.elements,
		Profiles:  this.infobase.profiles.elements,
		Matchers:  this.infobase.matchers.elements,
	}
	ipxe.RegisterHandler("/", kipxe.NewHandler(this.controller, "/", infobase))
	ipxe.Register("/ready", ready.Ready)

	ipxe.Start(nil, "", this.config.PXEPort)
	go func() {
		time.Sleep(2 * time.Second)
		ready.Register(&Ready{})
	}()

	if this.config.CacheDir != "" {
		this.controller.EnqueueCommand(CMD_CLEANUP)
	}
}

///////////////////////////////////////////////////////////////////////////////

func (this *reconciler) Config(cfg interface{}) *Config {
	return this.config
}

func (this *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	var err error
	logger.Infof("reconcile")
	switch obj.Data().(type) {
	case *v1alpha1.BootProfile:
		_, err = this.infobase.profiles.Update(logger, obj)
	case *v1alpha1.BootProfileMatcher:
		_, err = this.infobase.matchers.Update(logger, obj)
	case *v1alpha1.BootResource:
		_, err = this.infobase.resources.Update(logger, obj)
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
	case v1alpha1.RESOURCE:
		this.infobase.resources.Delete(logger, key.ObjectName())
	}
	return reconcile.Succeeded(logger)
}

func (this *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	this.infobase.cache.Cleanup(logger, this.config.CacheTTL)
	return reconcile.Succeeded(logger)
}
