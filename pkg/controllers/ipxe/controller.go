/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved.
 * This file is licensed under the Apache Software License, v. 2 except as noted
 * otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package ipxe

import (
	"path/filepath"
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"

	_apps "k8s.io/api/apps/v1"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/crds"
	api "github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

const NAME = "ipxe"

const CMD_CLEANUP = "cache-cleanup"

var secretGK = resources.NewGroupKind("", "Secret")

func init() {
	crds.AddToRegistry(apiextensions.DefaultRegistry())
}

func init() {
	_ = _apps.Deployment{}
	controller.Configure(NAME).
		RequireLease().
		Reconciler(Create).
		DefaultWorkerPool(5, 0).
		OptionsByExample("options", &Config{}).
		MainResourceByGK(api.MATCHER).
		CustomResourceDefinitions(api.MATCHER, api.PROFILE, api.DOCUMENT).
		WatchesByGK(api.PROFILE, api.DOCUMENT).
		WorkerPool(CMD_CLEANUP, 1, time.Minute).
		Commands(CMD_CLEANUP).
		MustRegister()
}

///////////////////////////////////////////////////////////////////////////////

func Create(controller controller.Interface) (reconcile.Interface, error) {
	var cache *kipxe.DirCache

	cfg, _ := controller.GetOptionSource("options")
	config := cfg.(*Config)
	if config.CacheDir != "" {
		path, err := filepath.Abs(config.CacheDir)
		if err != nil {
			return nil, err
		}
		cache, err = kipxe.NewDirectoryCache(controller, path)
		if err != nil {
			return nil, err
		}
	}

	this := &reconciler{
		controller: controller,
		config:     config,
		infobase:   GetSharedInfoBase(controller),
	}
	this.infobase.cache = cache
	return this, nil
}
