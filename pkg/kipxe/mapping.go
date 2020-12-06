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
	"github.com/mandelsoft/spiff/yaml"
)

type Mapping interface {
	Map(name string, values, metavalues simple.Values, intermediate Intermediate) (Intermediate, error)
}

type defaultMapping struct {
	SpiffTemplate
}

func NewDefaultMapping(m yaml.Node) Mapping {
	addImplicitAccess(m)
	return &defaultMapping{
		SpiffTemplate{m},
	}
}

func (this *defaultMapping) Map(name string, values, metavalues simple.Values, intermediate Intermediate) (Intermediate, error) {
	var err error

	intermediate = intermediate.Wrap()

	inputs := []yaml.Node{}
	err = this.AddStub(&inputs, fmt.Sprintf("%s:%s", name, "values"), values)
	if err != nil {
		return nil, err
	}
	err = this.AddStub(&inputs, fmt.Sprintf("%s:%s", name, "metadata"), metavalues)
	if err != nil {
		return nil, err
	}
	err = this.AddStub(&inputs, fmt.Sprintf("%s:%s", name, "intermediate"), intermediate)
	if err != nil {
		return nil, err
	}

	return this.MergeWith(inputs...)
}

func addImplicitAccess(m yaml.Node) {
	if vm, ok := m.Value().(map[string]yaml.Node); ok {
		if vm["metadata"] == nil {
			vm["metadata"] = yaml.NewNode("(( &temporary ))", "meta")
		}
		if vm["current"] == nil {
			vm["current"] = yaml.NewNode("(( &temporary ))", "meta")
		}
	}
}

func mapit(desc string, mapping Mapping, values, metavalues simple.Values, intermediate Intermediate) (Intermediate, error) {
	var err error
	if mapping != nil {
		if log {
			logger.Infof("mapping (by mapping) %s...", desc)
		}
		if !utils.IsNil(values) {
			n := simple.Values{}
			n["<<<"] = "(( merge ))"
			for k, v := range values {
				n[k] = v
			}
			values = n
		}
		intermediate, err = mapping.Map(desc, values, metavalues, intermediate)
		if err != nil {
			return nil, err
		}
	} else {
		if values != nil {
			if log {
				logger.Infof("mapping (by defaulting) %s...", desc)
			}
			return intermediate.Merge(values), nil
		}
		if log {
			logger.Infof("no mapping for %s", desc)
		}
	}
	return intermediate, nil
}
