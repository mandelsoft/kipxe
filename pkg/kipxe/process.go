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
	"encoding/json"
	"html/template"
	"strings"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"gopkg.in/yaml.v2"
)

func Process(name string, values simple.Values, src Source) (Source, error) {
	var data []byte
	var err error
	mime := src.MimeType()
	if strings.HasPrefix(mime, MIME_GTEXT) {
		mime = MIME_GTEXT
	}
	switch src.MimeType() {
	case MIME_JSON:
		in, err := src.Bytes()
		if err != nil {
			return nil, err
		}
		if in == nil {
			data, err = json.Marshal(values)
		} else {
			return src, nil
		}
	case MIME_YAML:
		in, err := src.Bytes()
		if err != nil {
			return nil, err
		}
		if in == nil {
			data, err = yaml.Marshal(values)
		} else {
			return src, nil
		}
	case MIME_TEXT, MIME_GTEXT, MIME_SHELL, MIME_XML:
		b, err := src.Bytes()
		if err != nil {
			return nil, err
		}
		logger.Infof("go template (len %d) with %s\n", len(b), values)
		t, err := template.New(name).Parse(string(b))
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
