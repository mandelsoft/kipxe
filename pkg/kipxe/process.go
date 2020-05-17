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
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"gopkg.in/yaml.v2"
)

const MIME_OCTET = restful.MIME_OCTET
const MIME_XML = restful.MIME_XML
const MIME_JSON = restful.MIME_JSON
const MIME_YAML = "application/x-yaml"
const MIME_TEXT = "text/plain"
const MIME_GTEXT = "text/"

type Source interface {
	MimeType() string
	Serve(w http.ResponseWriter, r *http.Request)
	Bytes() []byte
}

////////////////////////////////////////////////////////////////////////////////

type DataSource struct {
	mime string
	data []byte
}

func (this *DataSource) MimeType() string {
	return this.mime
}

func (this *DataSource) Bytes() []byte {
	return this.data
}

func (this *DataSource) Serve(w http.ResponseWriter, r *http.Request) {
	mime := this.MimeType()
	if mime != "" {
		w.Header().Add("Content-Type", mime)
	}
	w.Write(this.data)
}

func NewDataSource(mime string, data []byte) Source {
	return &DataSource{
		mime: mime,
		data: data,
	}
}

func NewTextSource(mime, text string) Source {
	logger.Infof("TXT: %s", text)
	return &DataSource{
		mime: mime,
		data: []byte(text),
	}
}

func NewBinarySource(mime, b64 string) (Source, error) {
	bytes := []byte(b64)
	l := base64.StdEncoding.DecodedLen(len(bytes))
	out := make([]byte, l, l)
	l, err := base64.StdEncoding.Decode(out, bytes)
	if err != nil {
		return nil, err
	}
	return NewDataSource(mime, out), nil
}

////////////////////////////////////////////////////////////////////////////////

type FilteredSource struct {
	DataSource
	source Source
}

func NewFilteredSource(src Source, data []byte) Source {
	return &FilteredSource{
		DataSource: DataSource{
			mime: src.MimeType(),
			data: data,
		},
		source: nil,
	}
}

////////////////////////////////////////////////////////////////////////////////

func Process(name string, values simple.Values, src Source) (Source, error) {
	var data []byte
	var err error
	mime := src.MimeType()
	if strings.HasPrefix(mime, MIME_GTEXT) {
		mime = MIME_GTEXT
	}
	switch src.MimeType() {
	case MIME_JSON:
		in := src.Bytes()
		if in == nil {
			data, err = json.Marshal(values)
		} else {
			return src, nil
		}
	case MIME_YAML:
		in := src.Bytes()
		if in == nil {
			data, err = yaml.Marshal(values)
		} else {
			return src, nil
		}
	case MIME_TEXT, MIME_GTEXT:
		logger.Infof("go template with %s\n%s", values, string(src.Bytes()))
		t, err := template.New(name).Parse(string(src.Bytes()))
		if err != nil {
			return nil, err
		}
		buf := &strings.Builder{}
		err = t.Execute(buf, values)
		if err != nil {
			return nil, err
		}
		data = []byte(buf.String())
	default:
		return src, nil
	}
	if err != nil {
		return nil, err
	}
	return NewFilteredSource(src, data), nil
}
