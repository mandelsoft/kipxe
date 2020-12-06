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
	"strings"
	"text/template"

	"github.com/emicklei/go-restful"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
)

const MIME_OCTET = restful.MIME_OCTET
const MIME_XML = restful.MIME_XML
const MIME_JSON = restful.MIME_JSON
const MIME_YAML = "application/x-yaml"
const MIME_CACERT = "application/x-x509-ca-cert"
const MIME_PEM = "application/x-pem-file"
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

type SourceMapper interface {
	Map(values simple.Values) (Source, error)
}

////////////////////////////////////////////////////////////////////////////////

func TemplateFor(name, txt string) (*template.Template, error) {
	if strings.Index(txt, "{{") < 0 {
		return nil, nil
	}
	t := template.New(name).Option("missingkey=error")
	return t.Parse(txt)
}

////////////////////////////////////////////////////////////////////////////////

type SourceSupport struct {
	mime string
}

func NewSourceSupport(mime string) SourceSupport {
	return SourceSupport{mime}
}

func (this *SourceSupport) MimeType() string {
	return this.mime
}

func (this *SourceSupport) IwriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func (this *SourceSupport) Iserve(data []byte, w http.ResponseWriter, r *http.Request) {
	mime := this.MimeType()
	if mime != "" {
		w.Header().Add(CONTENT_TYPE, mime)
	}
	w.Write(data)
}

////////////////////////////////////////////////////////////////////////////////

type DataSource struct {
	SourceSupport
	data []byte
}

func (this *DataSource) Bytes() ([]byte, error) {
	return this.data, nil
}

func NewNestedDataSource(mime string, data []byte) DataSource {
	return DataSource{
		SourceSupport: SourceSupport{mime},
		data:          data,
	}
}

func (this *DataSource) Serve(w http.ResponseWriter, r *http.Request) {
	this.Iserve(this.data, w, r)
}

func NewDataSource(mime string, data []byte) Source {
	return &DataSource{
		SourceSupport: SourceSupport{mime},
		data:          data,
	}
}

func NewTextSource(mime, text string) Source {
	logger.Infof("TXT: %s", text)
	return NewDataSource(mime, []byte(text))
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
	URLSource
}

var _ SourceMapper = &URLRedirectSource{}

func NewURLRedirectSource(src URLSource) URLSource {
	return &URLRedirectSource{
		URLSource: src,
	}
}

func (this *URLRedirectSource) Serve(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, this.URL(), 301)
}

func (this *URLRedirectSource) Map(values simple.Values) (Source, error) {
	if m, ok := this.URLSource.(SourceMapper); ok {
		mapped, err := m.Map(values)
		if err != nil {
			return nil, err
		}
		if m, ok := mapped.(URLSource); ok {
			return NewURLRedirectSource(m), nil
		}
		return mapped, nil
	}
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////

type URLSource interface {
	Source
	URL() string
	Cache() Cache
}

type urlSource struct {
	SourceSupport
	url   *url.URL
	cache Cache
}

func NewURLSource(mime string, url *url.URL, cache Cache) URLSource {
	return &urlSource{
		SourceSupport: SourceSupport{mime},
		url:           url,
		cache:         cache,
	}
}

func (this *urlSource) URL() string {
	return this.url.String()
}

func (this *urlSource) Cache() Cache {
	return this.cache
}

func (this *urlSource) Bytes() ([]byte, error) {
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

func (this *urlSource) Serve(w http.ResponseWriter, r *http.Request) {
	if this.cache != nil {
		this.cache.Serve(this.url, w, r)
		return
	}
	mime := this.MimeType()
	resp, err := http.Get(this.url.String())
	if err != nil {
		this.IwriteErrorResponse(w, http.StatusUnprocessableEntity, err)
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

type mappedURLSource struct {
	*urlSource
	url   string
	templ *template.Template
}

var _ SourceMapper = &mappedURLSource{}
var _ URLSource = &mappedURLSource{}

func NewMappedURLSource(mime string, rawURL string, cache Cache) (URLSource, error) {
	templ, err := TemplateFor("url", rawURL)
	if err != nil {
		return nil, err
	}
	if templ == nil {
		u, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("invalid url %q: ", rawURL, err)
		}
		return NewURLSource(mime, u, cache), nil
	}
	return &mappedURLSource{
		urlSource: &urlSource{
			SourceSupport: SourceSupport{mime},
			cache:         cache,
		},
		url:   rawURL,
		templ: templ,
	}, nil
}

func (this *mappedURLSource) URL() string {
	return this.url
}

func (this *mappedURLSource) Bytes() ([]byte, error) {
	return nil, fmt.Errorf("cannot serve url template")
}

func (this *mappedURLSource) Serve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write([]byte("cannot serve url template\n"))
}

func (this *mappedURLSource) Map(values simple.Values) (Source, error) {
	buf := &strings.Builder{}
	err := this.templ.Execute(buf, values)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(buf.String())
	if err != nil {
		return nil, fmt.Errorf("mapping result %q is no valid URL: %s", buf.String(), err)
	}
	return NewURLSource(this.MimeType(), url, this.Cache()), nil
}

////////////////////////////////////////////////////////////////////////////////

type FilteredSource struct {
	DataSource
	source Source
}

func NewFilteredSource(src Source, data []byte) Source {
	return &FilteredSource{
		DataSource: NewNestedDataSource(src.MimeType(), data),
		source:     src,
	}
}
