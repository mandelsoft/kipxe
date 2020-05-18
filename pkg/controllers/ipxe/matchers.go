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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type BootMatchers struct {
	ResourceCache
	elements *kipxe.BootProfileMatchers
}

func newMatchers(infobase *InfoBase) *BootMatchers {
	return &BootMatchers{
		ResourceCache: NewResourceCache(infobase, &v1alpha1.BootProfileMatcher{}),
		elements:      kipxe.NewMatchers(infobase.profiles.elements),
	}
}

func (this *BootMatchers) Setup(logger logger.LogContext) {
	if this.initialized {
		return
	}
	this.initialized = true
	if logger != nil {
		logger.Infof("setup matchers")
	}
	list, _ := this.resource.ListCached(labels.Everything())

	for _, l := range list {
		elem, err := this.Update(logger, l)
		if elem != nil {
			logger.Infof("found matcher %s", elem.Name())
		}
		if err != nil {
			logger.Infof("errorneous matcher %s: %s", l.GetName(), err)
		}
	}
}

func (this *BootMatchers) recheckUsers(users kipxe.NameSet) {
}

func (this *BootMatchers) Recheck(users kipxe.NameSet) {
	this.EnqueueAll(users, v1alpha1.MATCHER)
	this.elements.Recheck(users)
}

func (this *BootMatchers) Update(logger logger.LogContext, obj resources.Object) (*kipxe.BootProfileMatcher, error) {
	m, err := NewMatcher(obj.Data().(*v1alpha1.BootProfileMatcher))
	if err == nil {
		err = this.elements.Set(m)
	}
	if err != nil {
		logger.Errorf("invalid matcher: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*v1alpha1.BootProfileMatcher)
			mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*v1alpha1.BootProfileMatcher)
		mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_READY)
		mod.AssureStringValue(&m.Status.Message, "matcher ok")
		return nil
	})
	return m, err
}

func (this *BootMatchers) Delete(logger logger.LogContext, name resources.ObjectName) {
	this.elements.Delete(name)
}

func NewMatcher(m *v1alpha1.BootProfileMatcher) (*kipxe.BootProfileMatcher, error) {
	var err error
	sel := labels.Everything()
	if m.Spec.Selector != nil {
		sel, err = metav1.LabelSelectorAsSelector(m.Spec.Selector)
	}
	if err != nil {
		err = fmt.Errorf("%s", strings.Replace(err.Error(), " pod ", " profile ", -1))
		return nil, err
	}
	if m.Spec.Profile == "" {
		return nil, fmt.Errorf("no profile specified")
	}
	weight := 0
	if m.Spec.Weight != nil {
		weight = *m.Spec.Weight
	} else {
		if m.Spec.Selector != nil {
			weight = len(m.Spec.Selector.MatchExpressions) + len(m.Spec.Selector.MatchLabels)
		}
	}
	mapping, err := Compile("mapping", m.Spec.Mapping)
	if err != nil {
		return nil, err
	}
	return kipxe.NewMatcher(
		resources.NewObjectName(m.Namespace, m.Name),
		sel, mapping, m.Spec.Values.Values,
		resources.NewObjectName(m.Namespace, m.Spec.Profile),
		weight,
	), nil
}
