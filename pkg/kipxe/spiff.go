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

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/spiff/yaml"
)

var forcelog bool = false
var log bool = forcelog

func Trace(b bool) {
	log = b || forcelog
}

type SpiffTemplate struct {
	mapping yaml.Node
}

func (this *SpiffTemplate) AddStub(inp *[]yaml.Node, name string, v simple.Values) error {
	if v == nil {
		return nil
	}

	i, err := yaml.Sanitize(name, v)
	if err != nil {
		return fmt.Errorf("%s: invalid values: %s", name, err)
	}
	*inp = append(*inp, i)
	return nil
}

func (this *SpiffTemplate) MergeWith(inputs ...yaml.Node) (simple.Values, error) {
	ctx := spiffing.New().WithMode(0)
	stubs, err := ctx.PrepareStubs(inputs...)
	//outer := flow.NewProcessLocalEnvironment(nil, "mapper")
	//stubs, err := flow.PrepareStubs(outer, false, inputs...)
	if err != nil {
		return nil, err
	}
	if log {
		logger.Infof("-----------------------------------")
		for i, v := range append([]yaml.Node{this.mapping}, stubs...) {
			r, _ := yaml.Normalize(v)
			logger.Infof("<- %d: %s", i, simple.Values(r.(map[string]interface{})))
		}
	}
	result, err := ctx.ApplyStubs(this.mapping, stubs)
	//result, err := flow.Apply(nil, this.mapping, stubs, flow.Options{})
	if err != nil {
		return nil, err
	}
	v, err := yaml.Normalize(result)
	if err != nil {
		return nil, err
	}
	if log {
		logger.Infof("->: %s", simple.Values(v.(map[string]interface{})))
	}
	if log {
		logger.Infof("===================================")
	}
	m := v.(map[string]interface{})
	if out, ok := m["output"]; ok {
		if x, ok := out.(map[string]interface{}); ok {
			if log {
				logger.Infof("output ->: %s", simple.Values(x))
			}
			return simple.Values(x), nil
		}
		return nil, fmt.Errorf("unexpected type for mapping output")
	}
	if out, ok := m["metadata"]; ok {
		if x, ok := out.(map[string]interface{}); ok {
			if log {
				logger.Infof("meta ->: %s", simple.Values(x))
			}
			return simple.Values(x), nil
		}
		return nil, fmt.Errorf("unexpected type for mapping metadata")
	}
	if log {
		logger.Infof("all ->: %s", m)
	}
	return m, nil
}

func toBool(i interface{}) bool {
	if i == nil {
		return false
	}
	switch v := i.(type) {
	case bool:
		return v
	case string:
		return len(v) > 0
	case int64:
		return v != 0
	case map[string]interface{}:
		return len(v) > 0
	case []interface{}:
		return len(v) > 0
	default:
		return false
	}
}
