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
	"github.com/gardener/controller-manager-library/pkg/ctxutil"

	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

var infobaseKey = ctxutil.SimpleKey("infobase")

func GetSharedInfoBase(controller controller.Interface) *InfoBase {
	return controller.GetEnvironment().GetOrCreateSharedValue(infobaseKey, func() interface{} {
		return NewInfoBase(controller)
	}).(*InfoBase)
}

type InfoBase struct {
	controller controller.Interface
	cache      kipxe.Cache
	matchers   *Matchers
	profiles   *Profiles
	documents  *Documents
}

func NewInfoBase(controller controller.Interface) *InfoBase {
	b := &InfoBase{
		controller: controller,
	}

	b.documents = newDocuments(b)
	b.profiles = newProfiles(b)
	b.matchers = newMatchers(b)
	return b
}

func (this *InfoBase) Setup() {
	this.documents.Setup(this.controller)
	this.profiles.Setup(this.controller)
	this.matchers.Setup(this.controller)
}
