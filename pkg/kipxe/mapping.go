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
	"github.com/mandelsoft/spiff/yaml"
)

type Mapping interface {
	Map(name string, values, metavalues, intermediate simple.Values) (simple.Values, error)
}

type defaultMapping struct {
	SpiffTemplate
}

func NewDefaultMapping(m yaml.Node) Mapping {
	return &defaultMapping{
		SpiffTemplate{m},
	}
}

func (this *defaultMapping) Map(name string, values, metavalues, intermediate simple.Values) (simple.Values, error) {
	var err error

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
