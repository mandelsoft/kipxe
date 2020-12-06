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
	"github.com/gardener/controller-manager-library/pkg/utils"
	"github.com/mandelsoft/spiff/spiffing"
)

var forcelog bool = false
var log bool = forcelog

func Trace(b bool) {
	log = b || forcelog
}

type SpiffTemplate struct {
	mapping spiffing.Node
}

func (this *SpiffTemplate) AddStub(inp *[]spiffing.Node, name string, v interface{}) error {
	var err error

	if utils.IsNil(v) {
		return nil
	}

	var node spiffing.Node
	if i, ok := v.(Intermediate); ok {
		if i.IsNil() {
			return nil
		}
		node, err = i.Node()
		if err != nil {
			return err
		}
	} else {
		if i, ok := v.(spiffing.Node); ok {
			if i.Value() == nil {
				return nil
			}
			node = i
		} else {
			i, err := spiffing.ToNode(name, v)
			if err != nil {
				return fmt.Errorf("%s: invalid values: %s", name, err)
			}
			node = i
		}
	}
	*inp = append(*inp, node)
	return nil
}

func (this *SpiffTemplate) MergeWith(inputs ...spiffing.Node) (Intermediate, error) {
	ctx := spiffing.New().WithMode(0)
	if log {
		logger.Infof("-----------------------------------")
		for i, v := range append([]spiffing.Node{this.mapping}, inputs...) {
			r, _ := ctx.Normalize(v)
			logger.Infof("<- %d: %s", i, simple.Values(r.(map[string]interface{})))
		}
	}
	stubs, err := ctx.PrepareStubs(inputs...)
	if err != nil {
		return nil, err
	}
	result, err := ctx.ApplyStubs(this.mapping, stubs)
	//result, err := flow.Apply(nil, this.mapping, stubs, flow.Options{})
	if err != nil {
		return nil, err
	}
	i := NewNodeIntermediateValues(result)
	if log {
		v, _ := i.Values()
		logger.Infof("->: %s", v)
		logger.Infof("===================================")
	}
	m, err := i.Field("output")
	if m != nil {
		if log {
			v, _ := i.Values()
			logger.Infof("output ->: %s", v)
		}
		return m, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unexpected type for mapping output")
	}
	m, err = i.Field("metadata")
	if m != nil {
		if log {
			v, _ := i.Values()
			logger.Infof("meta ->: %s", v)
		}
		return m, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unexpected type for mapping metadata")
	}

	if log {
		logger.Infof("meta ->: use complete mapping result (see above)")
	}
	return i, nil
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
