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
	"sync"

	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
)

type BootResources struct {
	lock     sync.RWMutex
	elements map[string]*BootResource
	users    map[string]NameSet
}

func NewResources() *BootResources {
	return &BootResources{
		elements: map[string]*BootResource{},
		users:    map[string]NameSet{},
	}
}

func (this *BootResources) Recheck(set NameSet) NameSet {
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

func (this *BootResources) check(m *BootResource) error {
	return nil
}

func (this *BootResources) Get(name Name) *BootResource {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.elements[name.String()]
}

func (this *BootResources) Set(m *BootResource) NameSet {
	this.lock.Lock()
	defer this.lock.Unlock()

	key := m.Key()
	this.elements[key] = m
	m.error = this.check(m)

	users := this.users[key]
	if users == nil {
		return NameSet{}
	}
	return users.Copy()
}

func (this *BootResources) Delete(name Name) NameSet {
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

func (this *BootResources) AddUser(name Name, user Name) {
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

func (this *BootResources) AddUsersForAll(set NameSet, user Name) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, name := range set {
		key := name.String()
		set := this.users[key]
		if set == nil {
			set = NameSet{}
			this.users[key] = set
			set.Add(user)
		}
	}
}

func (this *BootResources) DeleteUser(name Name, user Name) {
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

func (this *BootResources) DeleteUsersForAll(set NameSet, user Name) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, name := range set {
		key := name.String()
		set := this.users[key]
		if set != nil {
			set.Remove(user)
			if len(set) == 0 {
				delete(this.users, key)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type BootResource struct {
	Element
	error          error
	values         simple.Values
	mapping        Mapping
	source         Source
	skipProcessing bool
}

func NewResource(name Name, mapping Mapping, values simple.Values, src Source, skipProcessing bool) *BootResource {
	return &BootResource{
		Element:        NewElement(name),
		source:         src,
		values:         values,
		mapping:        mapping,
		skipProcessing: skipProcessing,
	}
}

func (this *BootResource) GetSource() Source {
	return this.source
}

func (this *BootResource) GetValues() simple.Values {
	return this.values
}

func (this *BootResource) GetMapping() Mapping {
	return this.mapping
}

////////////////////////////////////////////////////////////////////////////////
