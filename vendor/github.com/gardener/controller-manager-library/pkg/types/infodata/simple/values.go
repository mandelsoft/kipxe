/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
 *
 *
 */

package simple

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/gardener/controller-manager-library/pkg/types/infodata"
)

const T_VALUES = infodata.TypeVersion("Values")
const T_VALUELIST = infodata.TypeVersion("ValueList")

func init() {
	infodata.Register(T_VALUES, infodata.UnmarshalFunc((Values)(nil)))
	infodata.Register(T_VALUELIST, infodata.UnmarshalFunc((ValueList)(nil)))
}

type Values map[string]interface{}

func (this Values) TypeVersion() infodata.TypeVersion {
	return T_VALUES
}

func (this Values) Marshal() ([]byte, error) {
	return json.Marshal(this)
}

func (this Values) String() string {
	b, _ := json.Marshal(this)
	return string(b)
}

type ValueList []interface{}

func (this ValueList) TypeVersion() infodata.TypeVersion {
	return T_VALUELIST
}

func (this ValueList) Marshal() ([]byte, error) {
	return json.Marshal(this)
}

///////////////////////////////////////////////////////////////////////////////

func (in Values) DeepCopy() Values {
	if in == nil {
		return nil
	}
	return runtime.DeepCopyJSON(in)
}

func (in ValueList) DeepCopy() ValueList {
	if in == nil {
		return nil
	}
	return runtime.DeepCopyJSONValue(in).([]interface{})
}

///////////////////////////////////////////////////////////////////////////////

func (this Values) Map(name string) (Values, bool) {
	if this == nil {
		return nil, true
	}
	m := this[name]
	if m == nil {
		return nil, true
	}
	v, ok := m.(map[string]interface{})
	return Values(v), ok
}

func (this Values) List(name string) (ValueList, bool) {
	if this == nil {
		return nil, true
	}
	m := this[name]
	if m == nil {
		return nil, true
	}
	v, ok := m.([]interface{})
	return ValueList(v), ok
}

func (this Values) StringValue(name string) (string, bool) {
	if this == nil {
		return "", false
	}
	m, ok := this[name]
	if !ok {
		return "", false
	}
	v, ok := m.(string)
	return v, ok
}

///////////////////////////////////////////////////////////////////////////////

func (this ValueList) Map(i int) (Values, bool) {
	if this == nil {
		return nil, true
	}
	if i >= len(this) || i < 0 {
		return nil, false
	}
	m := this[i]
	if m == nil {
		return nil, true
	}
	v, ok := m.(map[string]interface{})
	return Values(v), ok
}

func (this ValueList) List(i int) (ValueList, bool) {
	if this == nil {
		return nil, true
	}
	if i >= len(this) || i < 0 {
		return nil, false
	}
	m := this[i]
	if m == nil {
		return nil, true
	}
	v, ok := m.([]interface{})
	return ValueList(v), ok
}

func (this ValueList) StringValue(i int) (string, bool) {
	if this == nil {
		return "", false
	}
	if i >= len(this) || i < 0 {
		return "", false
	}
	m := this[i]
	v, ok := m.(string)
	return v, ok
}
