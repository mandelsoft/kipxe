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

	"github.com/gardener/controller-manager-library/pkg/convert"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
)

const MACHINE_FOUND = "MACHINE-FOUND"
const REQUEST_REJECT = "REQUEST-REJECT"

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

func merge(a, b simple.Values) simple.Values {
	for k, v := range b {
		if a[k] == nil {
			a[k] = v
		}
	}
	delete(a, "metadata")
	return a
}

func (this *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := this.serve(w, req)
	if err != nil {
		this.Error(err)
	}
}

func (this *Handler) requestMetadata(req *http.Request) (MetaData, string) {
	metadata := MetaData{}
	raw := req.URL.Query()

	path := req.URL.Path[len(this.path):]
	metadata["RESOURCE_PATH"] = path

	host := strings.Split(req.RemoteAddr, ":")[0]
	metadata["ORIGIN"] = host
	fill(metadata, raw)
	fill(metadata, req.Header)

	this.Infof("request %s: %s", path, metadata)
	return metadata, path
}

func (this *Handler) serve(w http.ResponseWriter, req *http.Request) error {
	var err error

	if !strings.HasPrefix(req.URL.Path, this.path) {
		return this.error(w, http.StatusNotFound, "invalid resource")
	}

	metadata, path := this.requestMetadata(req)

	if this.infobase.Registry != nil {
		metadata, err = this.infobase.Registry.Map(this, metadata, req)
		if err != nil {
			return this.error(w, http.StatusBadRequest, "cannot map metadata: %s", err)
		}
		if s := convert.BestEffortString(metadata[REQUEST_REJECT]); s != "" {
			return this.error(w, http.StatusNotAcceptable, "%s", s)
		}
	}

	this.Infof("matching %s", metadata)
	list := this.infobase.Matchers.Match(this, metadata)
	if len(list) == 0 {
		this.Infof("no matcher found")
		return this.error(w, http.StatusNotFound, "no matching matcher")
	}

	this.Infof("found %d matchers", len(list))

	for _, m := range list {
		pname := m.ProfileName()
		this.Infof("looking in matcher %s, profile %s", m.Key(), pname)
		profile := this.infobase.Profiles.Get(pname)
		if profile == nil {
			return this.error(w, http.StatusNotFound, "profile %q not found", pname)
		}

		d, list := profile.GetDeliverableForPath(path)
		if d == nil {
			continue
		}

		doc := this.infobase.Resources.Get(d.Name())
		if doc == nil {
			return this.error(w, http.StatusNotFound, "document %q for profile %q resource %q not found", d.Name(), pname, path)
		}

		this.Infof("found document %s in profile %s", d.Name(), pname)

		source := doc.GetSource()
		resmatch := types.CopyAndNormalize(list)
		metavalues := simple.Values{}
		metadata["<<<"] = "(( merge ))"
		if resmatch != nil {
			metadata["resource-match"] = resmatch
		}
		metavalues["metadata"] = simple.Values(metadata)
		intermediate := types.NormValues(metavalues)
		if !doc.skipProcessing {
			intermediate, err = mapit(fmt.Sprintf("matcher %s", m.Name()), m.GetMapping(), metavalues, m.GetValues(), intermediate)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}
			if resmatch != nil {
				intermediate["resource-match"] = resmatch
			}
			intermediate, err = mapit(fmt.Sprintf("profile %s", pname), profile.GetMapping(), metavalues, profile.GetValues(), intermediate)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}
			if resmatch != nil {
				intermediate["resource-match"] = resmatch
			}
			intermediate, err = mapit(fmt.Sprintf("profile %s, document %s", pname, d.Name()), doc.GetMapping(), metavalues, doc.GetValues(), intermediate)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}

			if m, ok := source.(SourceMapper); ok {
				source, err = m.Map(intermediate)
				if err != nil {
					return err
				}
			}

			source, err = Process("document", intermediate, source)
			if err != nil {
				return this.error(w, http.StatusUnprocessableEntity, err.Error())
			}
		}

		if m, ok := source.(SourceMapper); ok {
			source, err = m.Map(intermediate)
			if err != nil {
				return err
			}
		}
		source.Serve(w, req)
		return nil
	}
	return this.error(w, http.StatusNotFound, "no resource %q found in matches", path)
}

func fill(dst map[string]interface{}, src map[string][]string) {
	for k, l := range src {
		all := []interface{}{}
		for _, v := range l {
			if _, ok := dst[v]; !ok {
				dst[k] = v
			}
			all = append(all, v)
		}
		dst["__"+k+"__"] = all
	}
}
