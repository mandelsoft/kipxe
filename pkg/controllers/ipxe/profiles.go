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
	"fmt"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type Profiles struct {
	ResourceCache
	elements *kipxe.Profiles
}

func newProfiles(infobase *InfoBase) *Profiles {
	return &Profiles{
		ResourceCache: NewResourceCache(infobase, &v1alpha1.Profile{}),
		elements:      kipxe.NewProfiles(infobase.documents.elements),
	}
}

func (this *Profiles) Setup(logger logger.LogContext) {
	if this.initialized {
		return
	}
	this.initialized = true
	if logger != nil {
		logger.Infof("setup profiles")
	}
	list, _ := this.resource.ListCached(labels.Everything())

	for _, l := range list {
		matcher, err := this.Update(logger, l)
		if matcher != nil {
			logger.Infof("found profile %s", matcher)
		}
		if err != nil {
			logger.Infof("errorneous profile %s: %s", l.GetName(), err)
		}
	}
}

func (this *Profiles) recheckUsers(logger logger.LogContext, users kipxe.NameSet) {
	logger.Infof("found users: %s", users)
	this.matchers.Recheck(users)
}

func (this *Profiles) Recheck(users kipxe.NameSet) {
	this.EnqueueAll(users, v1alpha1.PROFILE)
	this.elements.Recheck(users)
}

func (this *Profiles) Update(logger logger.LogContext, obj resources.Object) (*kipxe.Profile, error) {
	m, err := NewProfile(obj.Data().(*v1alpha1.Profile))
	if err == nil {
		var users kipxe.NameSet
		users, err = this.elements.Set(m)
		this.recheckUsers(logger, users)
	}
	if err != nil {
		this.recheckUsers(logger, this.elements.Delete(obj.ObjectName()))
		logger.Errorf("invalid profile: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*v1alpha1.Profile)
			mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*v1alpha1.Profile)
		mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_READY)
		mod.AssureStringValue(&m.Status.Message, "profile ok")
		return nil
	})
	return m, err
}

func (this *Profiles) Delete(logger logger.LogContext, name resources.ObjectName) {
	this.recheckUsers(logger, this.elements.Delete(name))
}

func NewProfile(m *v1alpha1.Profile) (*kipxe.Profile, error) {
	name := resources.NewObjectName(m.Namespace, m.Name)
	deliverables := []*kipxe.Deliverable{}
	for i, r := range m.Spec.Resources {
		if strings.TrimSpace(r.DocumentName) == "" {
			return nil, fmt.Errorf("entry %d: empty document name", i)
		}
		deliverables = append(deliverables, kipxe.NewDeliverable(resources.NewObjectName(m.Namespace, r.DocumentName), r.Path))
	}

	mapping, err := Compile("mapping", m.Spec.Mapping)
	if err != nil {
		return nil, err
	}
	return kipxe.NewProfile(name, mapping, deliverables...)
}
