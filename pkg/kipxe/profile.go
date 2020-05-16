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

	"github.com/gardener/controller-manager-library/pkg/logger"
)

type Profiles struct {
	lock     sync.RWMutex
	elements map[string]*Profile
	nested   *Documents
	users    map[string]NameSet
}

func NewProfiles(nested *Documents) *Profiles {
	return &Profiles{
		elements: map[string]*Profile{},
		users:    map[string]NameSet{},
		nested:   nested,
	}
}

func (this *Profiles) Recheck(set NameSet) NameSet {
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

func (this *Profiles) check(m *Profile) error {
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

func (this *Profiles) Get(name Name) *Profile {
	this.lock.Lock()
	this.lock.Unlock()
	return this.elements[name.String()]
}

func (this *Profiles) Set(m *Profile) (NameSet, error) {
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

func (this *Profiles) Delete(name Name) NameSet {
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

func (this *Profiles) AddUser(name Name, user Name) {
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

func (this *Profiles) DeleteUser(name Name, user Name) {
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
	name Name
	path string
}

func NewDeliverable(name Name, path string) *Deliverable {
	return &Deliverable{name, path}
}

func (this *Deliverable) Name() Name {
	return this.name
}

func (this *Deliverable) Path() string {
	return this.path
}

////////////////////////////////////////////////////////////////////////////////

type Profile struct {
	Element
	error        error
	mapping      *Mapping
	deliverables map[string]*Deliverable
}

func NewProfile(name Name, mapping *Mapping, deliverables ...*Deliverable) (*Profile, error) {
	m := map[string]*Deliverable{}
	for i, d := range deliverables {
		if strings.TrimSpace(d.name.String()) == "" {
			return nil, fmt.Errorf("entry %d: empty document name", i)
		}
		if strings.TrimSpace(d.path) == "" {
			return nil, fmt.Errorf("entry %d: empty path", i)
		}
		if old := m[d.path]; old != nil {
			return nil, fmt.Errorf("duplicate deliverable for path %s (%s and %s)", old.name, d.name)
		}
		m[d.path] = d
	}
	return &Profile{
		Element:      NewElement(name),
		mapping:      mapping,
		deliverables: m,
	}, nil
}

func (this *Profile) Error() error {
	return this.error
}

func (this *Profile) Documents() NameSet {
	set := NameSet{}
	for _, d := range this.deliverables {
		set.Add(d.Name())
	}
	return set
}

////////////////////////////////////////////////////////////////////////////////
