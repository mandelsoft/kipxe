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
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/labels"
)

type Matchers struct {
	lock     sync.RWMutex
	elements map[string]*Matcher
	nested   *Profiles
}

func NewMatchers(nested *Profiles) *Matchers {
	return &Matchers{
		elements: map[string]*Matcher{},
		nested:   nested,
	}
}

func (this *Matchers) Recheck(set NameSet) NameSet {
	this.lock.Lock()
	this.lock.Unlock()
	recheck := NameSet{}
	for _, name := range set {
		e := this.elements[name.String()]
		if e.recheck(func() error { return this.check(e) }) {
			recheck.Add(name)
		}
	}
	return recheck
}

func (this *Matchers) check(m *Matcher) error {
	if e := this.nested.Get(m.profile); e != nil {
		if e.Error() != nil {
			return fmt.Errorf("profile %s: %s", e.Name(), e.Error())
		}
	} else {
		return fmt.Errorf("profile %s not found", m.profile)
	}
	return nil
}

func (this *Matchers) Set(m *Matcher) error {
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

func (this *Matchers) Delete(name Name) {
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

type Matcher struct {
	Element
	selector labels.Selector
	profile  Name
	weight   int
}

func NewMatcher(name Name, sel labels.Selector, profile Name, weight int) *Matcher {
	return &Matcher{
		Element:  NewElement(name),
		selector: sel,
		profile:  profile,
		weight:   weight,
	}
}

func (this Matcher) PreferOver(m *Matcher) bool {
	return this.Weight() < m.Weight() ||
		(this.Weight() == m.Weight() && strings.Compare(this.Key(), m.Key()) < 0)
}

func (this Matcher) Matches(labels labels.Labels) bool {
	return this.selector.Matches(labels)
}

func (this Matcher) Weight() int {
	return this.weight
}

func (this Matcher) ProfileName() Name {
	return this.profile
}

////////////////////////////////////////////////////////////////////////////////

func Match(labels labels.Labels, list []*Matcher) Name {
	var found *Matcher
	for _, m := range list {
		if found == nil || m.PreferOver(found) {
			if m.Matches(labels) {
				found = m
			}
		}
	}
	return found.ProfileName()
}
