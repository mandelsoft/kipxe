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
	"net/http"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type Labels simple.Values

var _ labels.Labels = Labels{}

func (this Labels) Has(key string) bool {
	if v, ok := this[key]; ok {
		if _, ok := v.(string); ok {
			return ok
		}
	}
	return false
}

func (this Labels) Get(key string) string {
	if v, ok := this[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type Handler struct {
	logger.LogContext
	path       string
	reconciler *reconciler
}

func NewHandler(path string, reconciler *reconciler) http.Handler {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return &Handler{
		LogContext: reconciler.controller.NewContext("server", "ipxe-server"),
		path:       path,
		reconciler: reconciler,
	}
}

func (this *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error

	values := simple.Values{}
	raw := req.URL.Query()

	if !strings.HasPrefix(req.URL.Path, this.path) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("invalid resourcer\n"))
		return
	}

	path := req.URL.Path[len(this.path):]
	for k, l := range raw {
		all := []interface{}{}
		for _, v := range l {
			values[k] = v
			all = append(all, v)
		}
		values["__"+k+"__"] = all
	}
	this.Infof("request %s: %s", path, values)

	list := this.reconciler.infobase.matchers.elements.Match(Labels(values))
	if len(list) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no matching matcher\n"))
		return
	}

	for _, m := range list {
		pname := m.ProfileName()
		logger.Infof("looking in matcher %s, profile %s", m.Key(), pname)
		profile := this.reconciler.infobase.profiles.elements.Get(pname)
		if profile == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("profile %q not found\n", pname)))
			return
		}

		d := profile.GetDeliverableForPath(path)
		if d == nil {
			continue
		}

		doc := this.reconciler.infobase.documents.elements.Get(d.Name())
		if doc == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("document %q for profile %q resource %q not found\n", d.Name(), pname, path)))
			return
		}

		logger.Infof("found document %s in profile %s", d.Name(), pname)
		intermediate := values.DeepCopy()
		mapping := profile.GetMapping()
		if mapping != nil {
			intermediate, err = mapping.Map(fmt.Sprintf("profile %s", pname), profile.GetValues(), intermediate)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(err.Error()))
				return
			}
		} else {
			if profile.GetValues() != nil {
				for k, v := range profile.GetValues() {
					if intermediate[k] == nil {
						intermediate[k] = v
					}
				}
			}
		}

		mapping = doc.GetMapping()
		if mapping != nil {
			intermediate, err = mapping.Map(fmt.Sprintf("profile %s, document %s", pname, d.Name()), profile.GetValues(), intermediate, values)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(err.Error()))
				return
			}
		} else {
			if doc.GetValues() != nil {
				for k, v := range profile.GetValues() {
					if intermediate[k] == nil {
						intermediate[k] = v
					}
				}
			}
		}

		source, err := kipxe.Process("document", intermediate, doc.GetSource())
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(err.Error()))
			return
		}

		source.Serve(w, req)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(fmt.Sprintf("no resource %q found in matches\n", path)))
	return
}
