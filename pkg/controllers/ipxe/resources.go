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
	"html/template"
	"net/url"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type BootResources struct {
	ResourceCache
	elements *kipxe.BootResources
}

func newResources(infobase *InfoBase) *BootResources {
	return &BootResources{
		ResourceCache: NewResourceCache(infobase, &v1alpha1.BootResource{}),
		elements:      kipxe.NewResources(),
	}
}

func (this *BootResources) Setup(logger logger.LogContext) {
	if this.initialized {
		return
	}
	this.initialized = true
	if logger != nil {
		logger.Infof("setup documents")
	}
	list, _ := this.resource.ListCached(labels.Everything())

	for _, l := range list {
		elem, err := this.Update(logger, l)
		if elem != nil {
			logger.Infof("found document %s", elem.Name())
		}
		if err != nil {
			logger.Infof("errorneous document %s: %s", l.GetName(), err)
		}
	}
}

func (this *BootResources) recheckUsers(logger logger.LogContext, users kipxe.NameSet) {
	logger.Infof("found users: %s", users)
	this.profiles.Recheck(users)
}

func (this *BootResources) Recheck(users kipxe.NameSet) {
	this.EnqueueAll(users, v1alpha1.RESOURCE)
	this.elements.Recheck(users)
}

func (this *BootResources) Update(logger logger.LogContext, obj resources.Object) (*kipxe.BootResource, error) {
	m, err := NewResource(obj.Data().(*v1alpha1.BootResource), this.InfoBase.cache)
	if err == nil {
		this.recheckUsers(logger, this.elements.Set(m))
	}
	if err != nil {
		this.recheckUsers(logger, this.elements.Delete(obj.ObjectName()))
		logger.Errorf("invalid document: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*v1alpha1.BootResource)
			mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*v1alpha1.BootResource)
		mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_READY)
		mod.AssureStringValue(&m.Status.Message, "document ok")
		return nil
	})
	return m, err
}

func (this *BootResources) Delete(logger logger.LogContext, name resources.ObjectName) {
	this.recheckUsers(logger, this.elements.Delete(name))
}

func NewResource(m *v1alpha1.BootResource, cache kipxe.Cache) (*kipxe.BootResource, error) {
	var source kipxe.Source
	mime := strings.TrimSpace(m.Spec.MimeType)
	if mime == "" {
		return nil, fmt.Errorf("mime type empty")
	}
	if m.Spec.Text != "" {
		_, err := template.New(m.Name).Parse(m.Spec.Text)
		if err != nil {
			return nil, fmt.Errorf("text is no valid go template: %s", err)
		}

		if m.Spec.Binary != "" || m.Spec.URL != "" {
			return nil, fmt.Errorf("Text cannot be combined with URL or Binary")
		}
		source = kipxe.NewTextSource(mime, m.Spec.Text)
	} else {
		if m.Spec.Binary != "" {
			s, err := kipxe.NewBinarySource(mime, m.Spec.Binary)
			if err != nil {
				return nil, fmt.Errorf("invalid binary data:%s", err)
			}
			source = s
		} else {
			if m.Spec.URL != "" {
				u, err := url.Parse(m.Spec.URL)
				if err != nil {
					return nil, fmt.Errorf("invalid URL (%s): %s", m.Spec.URL, err)
				}
				if m.Spec.Volatile {
					cache = nil
				}
				if m.Spec.Redirect != nil && *m.Spec.Redirect {
					source = kipxe.NewURLRedirectSource(mime, u, cache)
				} else {
					source = kipxe.NewURLSource(mime, u, cache)
				}
			} else {
				source = kipxe.NewDataSource(mime, nil)
			}
		}
	}
	mapping, err := Compile("mapping", m.Spec.Mapping)
	if err != nil {
		return nil, err
	}
	return kipxe.NewResource(resources.NewObjectName(m.Namespace, m.Name),
		mapping, m.Spec.Values.Values, source, m.Spec.Plain != nil && *m.Spec.Plain), nil
}
