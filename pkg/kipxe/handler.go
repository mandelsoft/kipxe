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
	"net/http"
	"strings"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

type MetaDataMapper interface {
	Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error)
}

type Registry struct {
	lock     sync.RWMutex
	registry []MetaDataMapper
}

var _ MetaDataMapper = &Registry{}

func NewRegistry() *Registry {
	return &Registry{}
}

func (this *Registry) Register(m MetaDataMapper) {
	if m != nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		this.registry = append(this.registry, m)
	}
}

func (this *Registry) Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	var err error

	for _, m := range this.registry {
		values, err = m.Map(logger, values, req)
		if err != nil {
			break
		}
	}
	return values, err
}

var registry = NewRegistry()

func RegisterMetaDataMapper(m MetaDataMapper) {
	registry.Register(m)
}

////////////////////////////////////////////////////////////////////////////////

type ErrorString string

func (e ErrorString) Error() string { return string(e) }

////////////////////////////////////////////////////////////////////////////////

type Handler struct {
	logger.LogContext
	path     string
	infobase *InfoBase
}

func NewHandler(logger logger.LogContext, path string, infobase *InfoBase) http.Handler {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return &Handler{
		LogContext: logger.NewContext("server", "ipxe-server"),
		path:       path,
		infobase:   infobase,
	}
}

func (this *Handler) error(w http.ResponseWriter, status int, msg string, args ...interface{}) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	w.WriteHeader(status)
	w.Write([]byte(msg + "\n"))
	return ErrorString(msg)
}

func merge(a, b simple.Values, set utils.StringSet) simple.Values {
	for k, v := range b {
		if mb, ok := v.(map[string]interface{}); ok && set != nil && a[k] != nil && set.Contains(k) {
			if ma, ok := a[k].(map[string]interface{}); ok {
				a[k] = map[string]interface{}(merge(ma, mb, nil))
				continue
			}
		}
		if a[k] == nil {
			a[k] = v
		}
	}
	return a
}

func (this *Handler) mapit(desc string, mapping Mapping, metavalues, values, intermediate simple.Values) (simple.Values, error) {
	var err error
	if mapping != nil {
		intermediate, err = mapping.Map(desc, values, metavalues, intermediate)
		if err != nil {
			return nil, err
		}
	} else {
		if values != nil {
			return merge(intermediate, values, utils.NewStringSet("metadata")), nil
		}
	}
	return intermediate, nil
}

func (this *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := this.serve(w, req)
	if err != nil {
		logger.Error(err)
	}
}

func (this *Handler) serve(w http.ResponseWriter, req *http.Request) error {
	var err error

	metadata := MetaData{}
	raw := req.URL.Query()

	if !strings.HasPrefix(req.URL.Path, this.path) {
		return this.error(w, http.StatusNotFound, "invalid resource")
	}

	path := req.URL.Path[len(this.path):]
	metadata["RESOURCE_PATH"] = path
	for k, l := range raw {
		all := []interface{}{}
		for _, v := range l {
			if _, ok := metadata[v]; !ok {
				metadata[k] = v
			}
			all = append(all, v)
		}
		metadata["__"+k+"__"] = all
	}
	this.Infof("request %s: %s", path, metadata)

	if this.infobase.Registry != nil {
		metadata, err = this.infobase.Registry.Map(this, metadata, req)
		if err != nil {
			return this.error(w, http.StatusBadRequest, "cannot map metadata: %s", err)
		}
	}
	list := this.infobase.Matchers.Match(metadata)
	if len(list) == 0 {
		logger.Infof("no matcher found")
		return this.error(w, http.StatusNotFound, "no matching matcher")
	}

	logger.Infof("found %d matchers", len(list))
	metavalues := simple.Values{}
	metadata["<<<"] = "(( merge ))"
	metavalues["metadata"] = simple.Values(metadata)

	for _, m := range list {
		pname := m.ProfileName()
		logger.Infof("looking in matcher %s, profile %s", m.Key(), pname)
		profile := this.infobase.Profiles.Get(pname)
		if profile == nil {
			return this.error(w, http.StatusNotFound, "profile %q not found", pname)
		}

		d := profile.GetDeliverableForPath(path)
		if d == nil {
			continue
		}

		doc := this.infobase.Resources.Get(d.Name())
		if doc == nil {
			return this.error(w, http.StatusNotFound, "document %q for profile %q resource %q not found", d.Name(), pname, path)
		}

		logger.Infof("found document %s in profile %s", d.Name(), pname)

		source := doc.GetSource()
		if !doc.skipProcessing {
			intermediate := metavalues
			intermediate, err = this.mapit(fmt.Sprintf("matcher %s", m.Name()), m.GetMapping(), metavalues, m.GetValues(), intermediate)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}
			intermediate, err = this.mapit(fmt.Sprintf("profile %s", pname), profile.GetMapping(), metavalues, profile.GetValues(), intermediate)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}
			intermediate, err = this.mapit(fmt.Sprintf("profile %s, document %s", pname, d.Name()), doc.GetMapping(), metavalues, doc.GetValues(), intermediate)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}

			source, err = Process("document", intermediate, source)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}
		}

		source.Serve(w, req)
		return nil
	}
	return this.error(w, http.StatusNotFound, "no resource %q found in matches", path)
}
