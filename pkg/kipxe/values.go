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
	"reflect"

	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
)

var mapType = reflect.TypeOf(map[string]interface{}{})
var arrayType = reflect.TypeOf([]interface{}{})

func NormValues(v simple.Values) simple.Values {
	return simple.Values(CopyAndNormalize(v).(map[string]interface{}))
}

func CopyAndNormalize(in interface{}) interface{} {
	if in == nil {
		return in
	}
	switch e := in.(type) {
	case map[string]string:
		r := map[string]interface{}{}
		for k, v := range e {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case map[string]interface{}:
		r := map[string]interface{}{}
		for k, v := range e {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case v1alpha1.Values:
		r := map[string]interface{}{}
		for k, v := range e.Values {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case simple.Values:
		r := map[string]interface{}{}
		for k, v := range e {
			r[k] = CopyAndNormalize(v)
		}
		return r
	case []interface{}:
		r := []interface{}{}
		for _, v := range e {
			r = append(r, v)
		}
		return r
	case []string:
		r := []interface{}{}
		for _, v := range e {
			r = append(r, v)
		}
		return r
	case string:
		return e
	case int:
		return int64(e)
	case int32:
		return int64(e)
	case float32:
		return float64(e)

	case int64, float64:
		return e
	default:
		value := reflect.ValueOf(e)
		if value.Kind() == reflect.Map {
			if value.Type().ConvertibleTo(mapType) {
				return CopyAndNormalize(value.Convert(mapType).Interface())
			}
			if value.Type().Key().Kind() == reflect.String {
				r := map[string]interface{}{}
				iter := value.MapRange()
				for iter.Next() {
					k := iter.Key()
					v := iter.Value()
					r[k.Interface().(string)] = CopyAndNormalize(v.Interface())
				}
				return r
			}
		}
		if value.Kind() == reflect.Array || value.Kind() == reflect.Slice {
			if value.Type().ConvertibleTo(arrayType) {
				return CopyAndNormalize(value.Convert(arrayType).Interface())
			}
			r := make([]interface{}, value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				r[i] = CopyAndNormalize(value.Index(i))
			}
			return r
		}
		panic(fmt.Errorf("invalid type %T", e))
	}
}
