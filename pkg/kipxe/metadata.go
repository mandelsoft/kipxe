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
	"strings"

	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"k8s.io/apimachinery/pkg/labels"
)

////////////////////////////////////////////////////////////////////////////////

type MetaData simple.Values

var _ labels.Labels = MetaData{}

func (this MetaData) String() string {
	return simple.Values(this).String()
}

func (this MetaData) DeepCopy() MetaData {
	return MetaData(simple.Values(this).DeepCopy())
}

func (this MetaData) Has(key string) bool {
	if v, ok := this[key]; ok {
		if _, ok := v.(string); ok {
			return ok
		}
	}
	comps := strings.Split(key, "/")
	if len(comps) > 1 {
		var data interface{}
		data = map[string]interface{}(this)
		for _, c := range comps {
			if m, ok := data.(map[string]interface{}); ok {
				data, ok = m[c]
				if !ok {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}

func (this MetaData) Get(key string) string {
	if v, ok := this[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	comps := strings.Split(key, "/")
	if len(comps) > 1 {
		var data interface{}
		data = map[string]interface{}(this)
		for _, c := range comps {
			if m, ok := data.(map[string]interface{}); ok {
				data, ok = m[c]
				if !ok {
					return ""
				}
			} else {
				return ""
			}
		}
		return fmt.Sprintf("%v", data)
	}
	return ""
}
