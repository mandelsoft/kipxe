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
	"net/url"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type MetaDataMappers struct {
	ResourceCache
	elements *kipxe.Registry
}

func newMappers(infobase *InfoBase) *MetaDataMappers {
	return &MetaDataMappers{
		ResourceCache: NewResourceCache(infobase, &v1alpha1.MetaDataMapper{}),
		elements:      infobase.registry,
	}
}

func (this *MetaDataMappers) Setup(logger logger.LogContext) {
	if this.initialized {
		return
	}
	this.initialized = true
	if logger != nil {
		logger.Infof("setup mappers")
	}
	list, _ := this.resource.ListCached(labels.Everything())

	for _, l := range list {
		elem, err := this.Update(logger, l)
		if elem != nil {
			logger.Infof("found mapper %s", elem.Name())
		}
		if err != nil {
			logger.Infof("errorneous mapper %s: %s", l.GetName(), err)
		}
	}
}

func (this *MetaDataMappers) recheckUsers(users kipxe.NameSet) {
}

func (this *MetaDataMappers) Recheck(users kipxe.NameSet) {
	this.EnqueueAll(users, v1alpha1.METADATAMAPPER)
}
func (this *MetaDataMappers) find(name resources.ObjectName) kipxe.MetaDataMapper {
	for _, m := range this.elements.Get() {
		if my, ok := m.(*MetaDataMapper); ok {
			if resources.EqualsObjectName(my.Name(), name) {
				return m
			}
		}
	}
	return nil
}

func (this *MetaDataMappers) Update(logger logger.LogContext, obj resources.Object) (*MetaDataMapper, error) {
	m, err := NewMapper(obj.Data().(*v1alpha1.MetaDataMapper))
	if err == nil {
		logger.Infof("update mapper registration")
		this.elements.SwitchRegistration(this.find(m.name), m)
	}
	if err != nil {
		logger.Errorf("invalid mapper: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*v1alpha1.MetaDataMapper)
			mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*v1alpha1.MetaDataMapper)
		mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_READY)
		mod.AssureStringValue(&m.Status.Message, "matcher ok")
		return nil
	})
	return m, err
}

func (this *MetaDataMappers) Delete(logger logger.LogContext, name resources.ObjectName) {
	for _, m := range this.elements.Get() {
		if my, ok := m.(*MetaDataMapper); ok {
			if resources.EqualsObjectName(my.Name(), name) {
				this.elements.Unregister(m)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type MetaDataMapper struct {
	kipxe.MetaDataMapper
	name resources.ObjectName
}

func (this *MetaDataMapper) Name() resources.ObjectName {
	return this.name
}

func (this *MetaDataMapper) String() string {
	return this.name.String()
}

func NewMapper(m *v1alpha1.MetaDataMapper) (*MetaDataMapper, error) {
	var mapper kipxe.MetaDataMapper
	name := resources.NewObjectName(m.Namespace, m.Name)

	if m.Spec.Mapping.Values != nil {
		if m.Spec.URL != nil {
			return nil, fmt.Errorf("multiple mapping options specified")
		}
		mapping, err := Compile(fmt.Sprintf("%s(mapping)", name), m.Spec.Mapping)
		if err != nil {
			return nil, fmt.Errorf("invalid mapping: %s", err)
		}
		mapper = kipxe.NewDefaultMetaDataMapper(mapping, m.Spec.Values.Values, m.Spec.Weight)
	} else {
		if m.Spec.URL != nil {
			u, err := url.Parse(*m.Spec.URL)
			if err != nil {
				return nil, fmt.Errorf("invalid URL: %s", err)
			}
			mapper = kipxe.NewURLMetaDataMapper(u, m.Spec.Weight)
		} else {
			return nil, fmt.Errorf("no mapping option specified")
		}
	}
	return &MetaDataMapper{
		mapper,
		name,
	}, nil
}
