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
	"net/http"

	"github.com/gardener/controller-manager-library/pkg/convert"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types"
	"github.com/onmetal/k8s-machines/pkg/machines"

	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type IndexMapper struct {
	index  machines.MachineIndex
	weight int
}

func NewIndexMapper(index machines.MachineIndex, weight int) kipxe.MetaDataMapper {
	return &IndexMapper{index, weight}
}

func (this *IndexMapper) Weight() int {
	return this.weight
}

func (this *IndexMapper) String() string {
	return "machine index mapper"
}

func (this *IndexMapper) Lookup(values kipxe.MetaData) *machines.Machine {
	uuid := values["uuid"]
	if uuid != nil {
		m := this.index.GetByUUID(uuid.(string))
		if m != nil {
			return m
		}
	}
	macs := values["__mac__"]
	if macs != nil {
		for _, v := range macs.([]interface{}) {
			m := this.index.GetByMAC(v.(string))
			if m != nil {
				return m
			}
		}
	}
	return nil
}

func (this *IndexMapper) Map(logger logger.LogContext, values kipxe.MetaData, req *http.Request) (kipxe.MetaData, error) {
	if convert.BestEffortBool(values[kipxe.MACHINE_FOUND]) {
		return values, nil
	}
	values = values.DeepCopy()
	m := this.Lookup(values)
	if m != nil {
		if m.UUID != "" {
			values["uuid"] = m.UUID
		}
		attrs := types.CopyAndNormalize(m.Values).(map[string]interface{})
		v, err := ObjectToValues(m.NICs, "nics")
		if err != nil {
			return nil, err
		}
		attrs["nics"] = types.CopyAndNormalize(v)
		values["attributes"] = attrs

		values[kipxe.MACHINE_FOUND] = true
		values["machine-name"] = m.Name.String()
		logger.Infof("found machine %s: %s", m.Name, values)
	} else {
		logger.Infof("no machine found")
	}
	return values, nil
}
