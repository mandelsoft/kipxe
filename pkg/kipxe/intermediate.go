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

	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/spiff/yaml"
)

type Intermediate interface {
	IsNil() bool
	Values() (simple.Values, error)
	Node() (spiffing.Node, error)
	FieldValue(name string) interface{}
	Field(name string) (Intermediate, error)
	Merge(b simple.Values) Intermediate
	Wrap() Intermediate
}

type intermediate struct {
	values simple.Values
	node   spiffing.Node
}

func NewSimpleIntermediateValues(v simple.Values) Intermediate {
	if utils.IsNil(v) {
		return &intermediate{nil, nil}
	}
	return &intermediate{v, nil}
}

func NewNodeIntermediateValues(v spiffing.Node) Intermediate {
	return &intermediate{nil, v}
}

func (this *intermediate) Values() (simple.Values, error) {
	var err error
	var v interface{}
	if this.IsNil() {
		return nil, nil
	}
	if this.values == nil {
		v, err = spiffing.Normalize(this.node)
		m, _ := v.(map[string]interface{})
		this.values = simple.Values(m)
	}
	return this.values, err
}

func (this *intermediate) Node() (spiffing.Node, error) {
	var err error
	if this.IsNil() {
		return nil, nil
	}
	if this.node == nil {
		this.node, err = spiffing.ToNode("intermediate", this.values)
	}
	return this.node, err
}

func (this *intermediate) Wrap() Intermediate {
	if this.node != nil {
		data := map[string]spiffing.Node{}
		for k, v := range this.node.Value().(map[string]spiffing.Node) {
			data[k] = v
		}
		data["current"] = this.node
		return NewNodeIntermediateValues(yaml.NewNode(data, "intermediate"))
	}
	data := simple.Values{}
	for k, v := range this.values {
		data[k] = v
	}
	data["current"] = this.values
	return NewSimpleIntermediateValues(data)
}

func (this *intermediate) FieldValue(name string) interface{} {
	if this.values != nil {
		return this.values[name]
	}
	if this.node != nil {
		if out, ok := this.node.Value().(map[string]spiffing.Node); ok {
			if out != nil && out[name] != nil && out[name].Value() != nil {
				r, _ := spiffing.Normalize(out[name])
				return r
			}
		}
	}
	return nil
}

func (this *intermediate) Field(name string) (Intermediate, error) {
	if this.node != nil {
		if out, ok := this.node.Value().(map[string]spiffing.Node); ok {
			if out != nil && out[name] != nil && out[name].Value() != nil {
				x := out[name]
				if _, ok := x.Value().(map[string]spiffing.Node); ok {
					return NewNodeIntermediateValues(x), nil
				} else {
					return nil, fmt.Errorf("unexpected type for field %q", name)
				}
			}
		}
		return nil, nil
	}
	if this.values != nil {
		if out, ok := this.values[name]; ok {
			if out != nil {
				if x, ok := out.(map[string]interface{}); ok {
					return NewSimpleIntermediateValues(simple.Values(x)), nil
				}
				return nil, fmt.Errorf("unexpected type for field %q", name)
			}
		}
		return nil, nil
	}
	return nil, nil
}

func (this *intermediate) IsNil() bool {
	if this.values != nil {
		return false
	}
	if this.node != nil {
		return this.node.Value() == nil
	}
	return true
}

func (this *intermediate) Merge(b simple.Values) Intermediate {
	if len(b) == 0 {
		return this
	}
	if this.IsNil() {
		return NewSimpleIntermediateValues(b.DeepCopy())
	}
	base := b.DeepCopy()

	cur, _ := this.Values()

	for k, v := range cur {
		if base[k] == nil {
			base[k] = v
		}
	}
	delete(base, "metadata")
	return NewSimpleIntermediateValues(base)
}
