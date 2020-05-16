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

func (this *Mapping) Map(name string, input, values simple.Values) (simple.Values, error) {
	inp := []yaml.Node{}
	if values != nil {
		i, err := yaml.Sanitize(name+":values", values)
		if err != nil {
			return nil, fmt.Errorf("invalid values: %s", err)
		}
		inp = append(inp, i)
	}
	i, err := yaml.Sanitize(name+":input", input)
	if err != nil {
		return nil, fmt.Errorf("invalid input: %s", err)
	}
	inp = append(inp, i)
	stubs, err := flow.PrepareStubs(nil, false, inp...)
	if err != nil {
		return nil, err
	}
	result, err := flow.Apply(nil, this.mapping, stubs)
	if err != nil {
		return nil, err
	}
	v, err := yaml.Normalize(result)
	if err != nil {
		return nil, err
	}
	return v.(map[string]interface{}), nil
}
