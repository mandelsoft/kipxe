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
)

type Name interface {
	String() string
}

////////////////////////////////////////////////////////////////////////////////

type DefaultName string

func (this DefaultName) String() string {
	return string(this)
}

type Named struct {
	name Name
	key  string
}

func NewNamed(name Name) Named {
	return Named{
		name: name,
		key:  name.String(),
	}
}

func (this *Named) Name() Name {
	return this.name
}

func (this *Named) Key() string {
	return this.key
}

////////////////////////////////////////////////////////////////////////////////

type NameSet map[string]Name

func NewNameSet(a ...Name) NameSet {
	return NameSet{}.Add(a...)
}

func NewnameSetByArray(a []Name) NameSet {
	s := NameSet{}
	if a != nil {
		s.Add(a...)
	}
	return s
}

func NewNameSetBySets(sets ...NameSet) NameSet {
	s := NameSet{}
	for _, set := range sets {
		for _, a := range set {
			s.Add(a)
		}
	}
	return s
}

func (this NameSet) String() string {
	sep := ""
	data := "["
	for k := range this {
		data = fmt.Sprintf("%s%s'%s'", data, sep, k)
		sep = ", "
	}
	return data + "]"
}

func (this NameSet) Clear() {
	for n := range this {
		delete(this, n)
	}
}

func (this NameSet) IsEmpty() bool {
	return len(this) == 0
}

func (this NameSet) Contains(n Name) bool {
	if n == nil {
		return false
	}
	_, ok := this[n.String()]
	return ok
}

func (this NameSet) Remove(n ...Name) NameSet {
	for _, p := range n {
		delete(this, p.String())
	}
	return this
}

func (this NameSet) AddAll(n []Name) NameSet {
	return this.Add(n...)
}

func (this NameSet) Add(n ...Name) NameSet {
	for _, p := range n {
		this[p.String()] = p
	}
	return this
}

func (this NameSet) AddSet(sets ...NameSet) NameSet {
	for _, s := range sets {
		for n, e := range s {
			this[n] = e
		}
	}
	return this
}

func (this NameSet) RemoveSet(sets ...NameSet) NameSet {
	for _, s := range sets {
		for n := range s {
			delete(this, n)
		}
	}
	return this
}

func (this NameSet) Equals(set NameSet) bool {
	for _, n := range set {
		if !this.Contains(n) {
			return false
		}
	}
	for _, n := range this {
		if !set.Contains(n) {
			return false
		}
	}
	return true
}

func (this NameSet) DiffFrom(set NameSet) (add, del NameSet) {
	add = NameSet{}
	del = NameSet{}
	for n, e := range set {
		if this[n] == nil {
			add[n] = e
		}
	}
	for n, e := range this {
		if set[n] == nil {
			del[n] = e
		}
	}
	return
}

func (this NameSet) Copy() NameSet {
	set := NewNameSet()
	for n, e := range this {
		set[n] = e
	}
	return set
}

func (this NameSet) Intersect(o NameSet) NameSet {
	set := NewNameSet()
	for n, e := range this {
		if o[n] != nil {
			set[n] = e
		}
	}
	return set
}

func (this NameSet) AsArray() []Name {
	a := []Name{}
	for _, e := range this {
		a = append(a, e)
	}
	return a
}
