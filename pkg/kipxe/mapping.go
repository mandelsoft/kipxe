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

type Mapping struct {
	mapping yaml.Node
}

func NewMapping(m yaml.Node) *Mapping {
	return &Mapping{
		mapping: m,
	}
}

var logMap = false

func (this *Mapping) Map(name string, values ...simple.Values) (simple.Values, error) {
	var err error
	var v interface{}

	if logMap {
		fmt.Printf("map %s\n", name)
		for i, s := range values {
			fmt.Printf("* stub %d:\n", i)
			b, _ := yaml2.Marshal(s)
			fmt.Printf("%s\n", string(b))
		}
		fmt.Printf("* template:\n")
		// v, err = yaml.Normalize(dynaml.ResetUnresolvedNodes(this.mapping))
		if err != nil {
			return nil, err
		}
		b, _ := yaml2.Marshal(v)
		fmt.Printf("%s\n", string(b))
	}

	inp := []yaml.Node{}
	for i, v := range values {
		if v != nil {
			i, err := yaml.Sanitize(fmt.Sprintf("%s:values[%d]", name, i), v)
			if err != nil {
				return nil, fmt.Errorf("invalid values: %s", err)
			}
			inp = append(inp, i)
		}
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
	return v.(map[string]interface{}), nil
}
