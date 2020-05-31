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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"github.com/mandelsoft/spiff/yaml"
)

type MetaDataMapper interface {
	Weight() int
	Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error)
}

type MetaDataMappers []MetaDataMapper

func (this MetaDataMappers) Len() int {
	return len(this)
}

func (this MetaDataMappers) Less(i, j int) bool {
	return this[i].Weight() < this[j].Weight()
}

func (this MetaDataMappers) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type Registry struct {
	lock     sync.RWMutex
	registry MetaDataMappers
	weight   int
}

var _ MetaDataMapper = &Registry{}

func NewRegistry(weight ...int) *Registry {
	w := 0
	for _, v := range weight {
		w += v
	}
	return &Registry{weight: w}
}

func (this *Registry) Has(m MetaDataMapper) int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.has(m)
}

func (this *Registry) has(m MetaDataMapper) int {
	for i, v := range this.registry {
		if m == v {
			return i
		}
	}
	return -1
}

func (this *Registry) Register(m MetaDataMapper) {
	if m != nil {
		if this.has(m) < 0 {
			this.lock.Lock()
			defer this.lock.Unlock()
			this.registry = append(this.registry, m)
			sort.Sort(sort.Reverse(this.registry))
		}
	}
}

func (this *Registry) Unregister(m MetaDataMapper) {
	if m != nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		if i := this.has(m); i >= 0 {
			this.registry = append(this.registry[:i], this.registry[i+1:]...)
		}
	}
}

func (this *Registry) SwitchRegistration(old, new MetaDataMapper) {
	if new != nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		if i := this.has(old); i >= 0 {
			this.registry = append(this.registry[:i], this.registry[i+1:]...)
		}
		if i := this.has(old); i < 0 {
			this.registry = append(this.registry, new)
			sort.Sort(sort.Reverse(this.registry))
		}
	} else {
		this.Unregister(old)
	}
}

func (this *Registry) Get() MetaDataMappers {
	this.lock.RLock()
	defer this.lock.RUnlock()
	result := make(MetaDataMappers, len(this.registry), len(this.registry))
	for i, v := range this.registry {
		result[i] = v
	}
	return result
}

func (this *Registry) Weight() int {
	return this.weight
}

func (this *Registry) Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	var err error

	logger.Infof("found %d metadata mappers", len(this.registry))
	for _, m := range this.registry {
		logger.Infof("  mapping metadata with %s", m)
		values, err = m.Map(logger, values, req)
		if err != nil {
			logger.Errorf("mapping failed: %s", err)
			break
		}
		logger.Infof("mapped to: %s", values)
	}
	return values, err
}

var registry = NewRegistry()

func RegisterMetaDataMapper(m MetaDataMapper) {
	registry.Register(m)
}

////////////////////////////////////////////////////////////////////////////////

type defaultMapper struct {
	mapping Mapping
	values  simple.Values
	weight  int
}

var _ MetaDataMapper = &defaultMapper{}

func NewDefaultMetaDataMapper(m yaml.Node, values simple.Values, weight int) MetaDataMapper {
	addMetadataAccess(m)
	return &defaultMapper{
		NewDefaultMapping(m),
		values,
		weight,
	}
}

func (this *defaultMapper) Weight() int {
	return this.weight
}

func (this *defaultMapper) Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error) {
	inp := simple.Values{}
	inp["metadata"] = NormValues(simple.Values(values))

	r, err := mapit("metadata", this.mapping, inp, this.values, simple.Values(values))
	if err != nil {
		return nil, err
	}
	return MetaData(r), nil
}

////////////////////////////////////////////////////////////////////////////////

type urlMapper struct {
	weight int
	url    *url.URL
}

var _ MetaDataMapper = &urlMapper{}

func NewURLMetaDataMapper(url *url.URL, weight int) MetaDataMapper {
	return &urlMapper{
		weight,
		url,
	}
}
func (this *urlMapper) Weight() int {
	return this.weight
}

func (this *urlMapper) Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	r, err := http.NewRequest("POST", this.url.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	r.Header.Set(CONTENT_TYPE, MIME_JSON)
	r.Header.Set("Accept", MIME_JSON)
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.Header.Get(CONTENT_TYPE) != MIME_JSON {
		return nil, fmt.Errorf("unexpected content type %s", resp.Header.Get(CONTENT_TYPE))
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return MetaData(result), nil
}
