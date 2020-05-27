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

package kipxe

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"k8s.io/apimachinery/pkg/labels"
)

type BootProfileMatchers struct {
	lock     sync.RWMutex
	elements map[string]*BootProfileMatcher
	nested   *BootProfiles
}

func NewMatchers(nested *BootProfiles) *BootProfileMatchers {
	return &BootProfileMatchers{
		elements: map[string]*BootProfileMatcher{},
		nested:   nested,
	}
}

func (this *BootProfileMatchers) Recheck(set NameSet) NameSet {
	this.lock.RLock()
	defer this.lock.RUnlock()
	recheck := NameSet{}
	for _, name := range set {
		e := this.elements[name.String()]
		if e != nil && e.recheck(func() error { return this.check(e) }) {
			recheck.Add(name)
		}
	}
	return recheck
}

func (this *BootProfileMatchers) check(m *BootProfileMatcher) error {
	if e := this.nested.Get(m.profile); e != nil {
		if e.Error() != nil {
			return fmt.Errorf("profile %s: %s", e.Name(), e.Error())
		}
	} else {
		return fmt.Errorf("profile %s not found", m.profile)
	}
	return nil
}

func (this *BootProfileMatchers) Get(name Name) *BootProfileMatcher {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.elements[name.String()]
}

func (this *BootProfileMatchers) Set(m *BootProfileMatcher) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := m.Key()
	old := this.elements[key]
	if old != nil {
		if old.profile.String() != m.profile.String() {
			this.nested.DeleteUser(old.ProfileName(), m.Name())
		}
	}
	this.nested.AddUser(m.ProfileName(), m.Name())
	this.elements[key] = m
	m.error = this.check(m)
	return m.error
}

func (this *BootProfileMatchers) Delete(name Name) {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := name.String()
	old := this.elements[key]
	if old != nil {
		delete(this.elements, key)
		this.nested.Delete(old.profile)
	}
}

////////////////////////////////////////////////////////////////////////////////

type BootProfileMatcher struct {
	Element
	selector labels.Selector
	matcher  Mapping
	mapping  Mapping
	values   simple.Values
	profile  Name
	weight   int
}

func NewMatcher(name Name, sel labels.Selector, matcher, mapping Mapping, values simple.Values, profile Name, weight int) (*BootProfileMatcher, error) {
	return &BootProfileMatcher{
		Element:  NewElement(name),
		selector: sel,
		matcher:  matcher,
		mapping:  mapping,
		values:   values,
		profile:  profile,
		weight:   weight,
	}, nil
}

func (this BootProfileMatcher) PreferOver(m *BootProfileMatcher) bool {
	return this.Weight() > m.Weight() ||
		(this.Weight() == m.Weight() && strings.Compare(this.Key(), m.Key()) < 0)
}

func (this BootProfileMatcher) Matches(logger logger.LogContext, meta MetaData) bool {
	if !this.selector.Matches(meta) {
		return false
	}
	if this.matcher != nil {
		metavalues := simple.Values{"metadata": simple.Values(meta)}
		r, err := this.matcher.Map("matcher", this.values, metavalues, nil)
		if err != nil {
			logger.Errorf("matcher %s failed: %s", this.Name(), err)
			return false
		}
		if m, ok := r["match"]; ok {
			//logger.Infof("matcher %s: %v", this.name, m)
			return toBool(m)
		}
		return false
	}
	return true
}

func (this *BootProfileMatcher) GetMapping() Mapping {
	return this.mapping
}

func (this *BootProfileMatcher) GetValues() simple.Values {
	return this.values
}

func (this BootProfileMatcher) Weight() int {
	return this.weight
}

func (this BootProfileMatcher) ProfileName() Name {
	return this.profile
}

////////////////////////////////////////////////////////////////////////////////

type BootProfileMatcherSlice []*BootProfileMatcher

var _ sort.Interface = BootProfileMatcherSlice{}

func (s BootProfileMatcherSlice) Len() int {
	return len(s)
}

func (s BootProfileMatcherSlice) Less(i, j int) bool {
	return s[i].PreferOver(s[j])
}

func (s BootProfileMatcherSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (this *BootProfileMatchers) Match(logger logger.LogContext, meta MetaData) BootProfileMatcherSlice {
	this.lock.RLock()
	defer this.lock.RUnlock()

	var found []*BootProfileMatcher
	for _, m := range this.elements {
		if m.Matches(logger, meta) {
			found = append(found, m)
		}
	}
	sort.Sort(BootProfileMatcherSlice(found))
	return found
}
