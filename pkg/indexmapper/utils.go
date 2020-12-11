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

package indexmapper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/utils"
)

func ObjectToValues(o interface{}, kind string) (interface{}, error) {
	if utils.IsNil(o) {
		return nil, nil
	}
	var values interface{}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("error marshalling %s: %s", kind, err)
	}
	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		values = []interface{}{}
	} else {
		values = map[string]interface{}{}
	}
	err = json.Unmarshal(data, &values)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling %s: %s", kind, err)
	}
	return values, nil
}
