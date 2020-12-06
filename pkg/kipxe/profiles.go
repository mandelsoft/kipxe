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
	"regexp"
	"strings"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
)

type BootProfiles struct {
	lock     sync.RWMutex
	elements map[string]*BootProfile
	nested   *BootResources
	users    map[string]NameSet
}

func NewProfiles(nested *BootResources) *BootProfiles {
	return &BootProfiles{
		elements: map[string]*BootProfile{},
		users:    map[string]NameSet{},
		nested:   nested,
	}
}

func (this *BootProfiles) Recheck(set NameSet) NameSet {
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

func (this *BootProfiles) check(m *BootProfile) error {
	for _, d := range m.deliverables {
		if e := this.nested.Get(d.Name()); e != nil {
			if e.Error() != nil {
				return fmt.Errorf("document %s: %s", d.Name(), e.Error())
			}
		} else {
			return fmt.Errorf("document %s not found", d.Name())
		}
	}
	return nil
}

func (this *BootProfiles) Get(name Name) *BootProfile {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.elements[name.String()]
}

func (this *BootProfiles) Set(m *BootProfile) (NameSet, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := m.Key()
	add := m.Documents()
	logger.Infof("documents for profile %s: %s", key, add)
	old := this.elements[key]
	if old != nil {
		oldd := old.Documents()
		var del NameSet
		add, del = oldd.DiffFrom(add)
		this.nested.DeleteUsersForAll(del, m.Name())
	}
	this.elements[m.Key()] = m
	this.nested.AddUsersForAll(add, m.Name())

	m.error = this.check(m)

	users := this.users[key]
	if users == nil {
		return NameSet{}, m.error
	}
	return users.Copy(), m.error
}

func (this *BootProfiles) Delete(name Name) NameSet {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := name.String()
	old := this.elements[key]
	if old != nil {
		delete(this.elements, key)
	}
	users := this.users[key]
	if users == nil {
		return NameSet{}
	}
	return users.Copy()
}

func (this *BootProfiles) AddUser(name Name, user Name) {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := name.String()
	set := this.users[key]
	if set == nil {
		set = NameSet{}
		this.users[key] = set
	}
	set.Add(user)
}

func (this *BootProfiles) DeleteUser(name Name, user Name) {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := name.String()
	set := this.users[key]
	if set != nil {
		set.Remove(user)
		if len(set) == 0 {
			delete(this.users, key)

		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type Deliverable struct {
	name    Name
	path    string
	pattern *regexp.Regexp
}

func NewDeliverable(name Name, path string) *Deliverable {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return &Deliverable{name, path, nil}
}

func NewDeliverableByPattern(name Name, pattern string) (*Deliverable, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &Deliverable{name, "", re}, nil
}

func (this *Deliverable) Name() Name {
	return this.name
}

func (this *Deliverable) Path() string {
	return this.path
}

func (this *Deliverable) Pattern() *regexp.Regexp {
	return this.pattern
}

////////////////////////////////////////////////////////////////////////////////

type BootProfile struct {
	Element
	error        error
	deliverables map[string]*Deliverable
	patterns     []*Deliverable
	paths        map[string]*Deliverable
}

func NewProfile(name Name, mapping Mapping, values simple.Values, deliverables ...*Deliverable) (*BootProfile, error) {
	m := map[string]*Deliverable{}
	paths := map[string]*Deliverable{}
	patterns := []*Deliverable{}
	for i, d := range deliverables {
		if strings.TrimSpace(d.name.String()) == "" {
			return nil, fmt.Errorf("entry %d: empty document name", i)
		}
		if strings.TrimSpace(d.path) == "" {
			if d.pattern == nil {
				return nil, fmt.Errorf("entry %d: neither path nor pattern specified", i)
			}
			patterns = append(patterns, d)
		} else {
			if d.pattern != nil {
				return nil, fmt.Errorf("entry %d: both, path and pattern specified", i)
			}
			if old := m[d.path]; old != nil {
				return nil, fmt.Errorf("duplicate deliverable for path %s (%s and %s)", old.name, d.name)
			}
			paths[d.path] = d
		}
		m[d.name.String()] = d
	}
	return &BootProfile{
		Element:      NewElement(name, values, mapping),
		deliverables: m,
		paths:        paths,
		patterns:     patterns,
	}, nil
}

func (this *BootProfile) Error() error {
	return this.error
}

func (this *BootProfile) Documents() NameSet {
	set := NameSet{}
	for _, d := range this.deliverables {
		set.Add(d.Name())
	}
	return set
}

func (this *BootProfile) GetMapping() Mapping {
	return this.mapping
}

func (this *BootProfile) GetValues() simple.Values {
	return this.values
}

func (this *BootProfile) GetDeliverableForPath(path string) (*Deliverable, []string) {
	d := this.paths[path]
	if d != nil {
		return d, []string{path}
	}
	for _, p := range this.patterns {
		if list := p.pattern.FindStringSubmatch(path); len(list) > 0 {
			if list[0] == path || "/"+list[0] == path {
				return p, list
			}
		}
	}
	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////
