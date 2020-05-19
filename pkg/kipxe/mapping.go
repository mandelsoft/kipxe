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
	yaml2 "github.com/ghodss/yaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

type Mapping interface {
	Map(name string, values, metavalues, intermediate simple.Values) (simple.Values, error)
}

type defaultMapping struct {
	mapping yaml.Node
}

func NewDefaultMapping(m yaml.Node) Mapping {
	return &defaultMapping{
		mapping: m,
	}
}

var logMap = false

func (this *defaultMapping) add(inp *[]yaml.Node, name string, v simple.Values) error {
	if v == nil {
		return nil
	}
	if logMap {
		fmt.Printf("* stub %s\n", name)
		b, _ := yaml2.Marshal(v)
		fmt.Printf("%s\n", string(b))
	}

	i, err := yaml.Sanitize(name, v)
	if err != nil {
		return fmt.Errorf("%s: invalid values: %s", name, err)
	}
	*inp = append(*inp, i)
	return nil
}

func (this *defaultMapping) Map(name string, values, metavalues, intermediate simple.Values) (simple.Values, error) {
	var err error
	var v interface{}

	if logMap {
		fmt.Printf("map %s\n", name)
		fmt.Printf("* template:\n")
		// v, err = yaml.Normalize(dynaml.ResetUnresolvedNodes(this.mapping))
		if err != nil {
			return nil, err
		}
		b, _ := yaml2.Marshal(v)
		fmt.Printf("%s\n", string(b))
	}

	inp := []yaml.Node{}
	err = this.add(&inp, fmt.Sprintf("%s:%s", name, "values"), values)
	if err != nil {
		return nil, err
	}
	err = this.add(&inp, fmt.Sprintf("%s:%s", name, "metadata"), metavalues)
	if err != nil {
		return nil, err
	}
	err = this.add(&inp, fmt.Sprintf("%s:%s", name, "intermediate"), intermediate)
	if err != nil {
		return nil, err
	}

	stubs, err := flow.PrepareStubs(nil, false, inp...)
	if err != nil {
		return nil, err
	}
	result, err := flow.Apply(nil, this.mapping, stubs)
	if err != nil {
		return nil, err
	}
	v, err = yaml.Normalize(result)
	if err != nil {
		return nil, err
	}
	if logMap {
		fmt.Printf("* result:\n")
		b, _ := yaml2.Marshal(v)
		fmt.Printf("%s\n", string(b))
	}

	m := v.(map[string]interface{})
	if out, ok := m["output"]; ok {
		if v, ok := out.(map[string]interface{}); ok {
			return simple.Values(v), nil
		}
		return nil, fmt.Errorf("unexpected type for mapping output")
	}
	return m, nil
}
