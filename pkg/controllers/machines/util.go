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
	"github.com/gardener/controller-manager-library/pkg/resources"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type ResourceCache struct {
	controller  controller.Interface
	resource    resources.Interface
	initialized bool
}

func NewResourceCache(controller controller.Interface, proto resources.ObjectData) ResourceCache {
	resc, _ := controller.GetMainCluster().Resources().Get(proto)
	return ResourceCache{
		controller: controller,
		resource:   resc,
	}
}

func (this *ResourceCache) ClusterKey(name resources.ObjectName, gk ...schema.GroupKind) resources.ClusterObjectKey {
	if len(gk) > 0 {
		return resources.NewClusterKey(this.controller.GetMainCluster().GetId(), gk[0], name.Namespace(), name.Name())
	}
	return resources.NewClusterKey(this.controller.GetMainCluster().GetId(), this.resource.GroupKind(), name.Namespace(), name.Name())
}

func (this *ResourceCache) Enqueue(name resources.ObjectName, gk ...schema.GroupKind) {
	this.controller.EnqueueKey(this.ClusterKey(name, gk...))
}

func (this *ResourceCache) EnqueueAll(set kipxe.NameSet, gk ...schema.GroupKind) {
	for _, u := range set {
		if o, ok := u.(resources.ObjectName); ok {
			this.Enqueue(o, gk...)
		}
	}
}
