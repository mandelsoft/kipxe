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
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/emicklei/go-restful"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

const MIME_OCTET = restful.MIME_OCTET
const MIME_XML = restful.MIME_XML
const MIME_JSON = restful.MIME_JSON
const MIME_YAML = "application/x-yaml"
const MIME_SHELL = "application/x-sh"
const MIME_TEXT = "text/plain"
const MIME_GTEXT = "text/"

const CONTENT_TYPE = "Content-Type"
const CONTENT_URL = "URL"

type Source interface {
	MimeType() string
	Serve(w http.ResponseWriter, r *http.Request)
	Bytes() ([]byte, error)
}

////////////////////////////////////////////////////////////////////////////////

type DataSource struct {
	mime string
	data []byte
}

func (this *DataSource) MimeType() string {
	return this.mime
}

func (this *DataSource) Bytes() ([]byte, error) {
	return this.data, nil
}

func (this *DataSource) Serve(w http.ResponseWriter, r *http.Request) {
	mime := this.MimeType()
	if mime != "" {
		w.Header().Add(CONTENT_TYPE, mime)
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

type URLRedirectSource struct {
	*URLSource
}

func NewURLRedirectSource(mime string, url *url.URL, cache Cache) Source {
	return &URLRedirectSource{
		URLSource: NewURLSource(mime, url, cache),
	}
}

func (this *URLRedirectSource) Serve(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, this.url.String(), 301)
}

////////////////////////////////////////////////////////////////////////////////

type URLSource struct {
	mime  string
	url   *url.URL
	cache Cache
}

func NewURLSource(mime string, url *url.URL, cache Cache) *URLSource {
	return &URLSource{
		mime:  mime,
		url:   url,
		cache: cache,
	}
}

func (this *URLSource) MimeType() string {
	return this.mime
}

func (this *URLSource) Bytes() ([]byte, error) {
	if this.cache != nil {
		return this.cache.Bytes(this.url)
	}
	resp, err := http.Get(this.url.String())
	if err != nil {
		return nil, fmt.Errorf("URL get failed: %s", err)
	}
	defer resp.Body.Close()
	buf := bytes.Buffer{}
	var tmp [8196]byte

	for {
		n, err := resp.Body.Read(tmp[:])
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if n < 0 {
			break
		}
	}
	return buf.Bytes(), nil
}

func (this *URLSource) Serve(w http.ResponseWriter, r *http.Request) {
	if this.cache != nil {
		this.cache.Serve(this.url, w, r)
		return
	}
	mime := this.MimeType()
	resp, err := http.Get(this.url.String())
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	t := resp.Header.Get(CONTENT_TYPE)
	if t != "" {
		mime = t
	}
	if mime != "" {
		w.Header().Add(CONTENT_TYPE, mime)
	}
	defer resp.Body.Close()
	var tmp [8196]byte

	for {
		n, err := resp.Body.Read(tmp[:])
		if n > 0 {
			w.Write(tmp[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if n < 0 {
			break
		}
	}
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
