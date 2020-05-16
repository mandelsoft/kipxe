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

package main

import (
	"encoding/json"
	"fmt"

	values "github.com/gardener/controller-manager-library/pkg/types/infodata/simple"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
)

func main() {
	v := values.Values{
		"v": map[string]interface{}{
			"x": 1,
			"z": 32.2,
		},
		"a": []interface{}{
			"v1",
		},
	}

	V := v1alpha1.Values{v1alpha1.NormValues(v)}
	s, _ := json.MarshalIndent(V.DeepCopy(), "", "  ")

	fmt.Printf("%s\n", s)

	u := &v1alpha1.Values{}
	err := json.Unmarshal(s, u)
	if err != nil {
		panic(err)
	}
	fmt.Printf("=== %v\n", u)

}
