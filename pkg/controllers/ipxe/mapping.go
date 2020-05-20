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

package ipxe

import (
	"fmt"
	"sort"

	"github.com/mandelsoft/spiff/compile"
	"github.com/mandelsoft/spiff/yaml"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

func Compile(field string, values v1alpha1.Values) (yaml.Node, error) {
	if values.Values == nil {
		return nil, nil
	}
	mapping, errs := compile.Compile(field, values.Values)
	if errs != nil {
		sort.Sort(errs)
		return nil, fmt.Errorf("error in %s: %s", field, errs)
	}
	return mapping, nil
}

func Mapping(field string, values v1alpha1.Values) (kipxe.Mapping, error) {
	if values.Values == nil {
		return nil, nil
	}
	node, err := Compile(field, values)
	if err != nil {
		return nil, err
	}
	return kipxe.NewDefaultMapping(node), nil
}
