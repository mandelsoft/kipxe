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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type BootResources struct {
	ResourceCache
	elements *kipxe.BootResources
}

func newResources(infobase *InfoBase) *BootResources {
	return &BootResources{
		ResourceCache: NewResourceCache(infobase, &v1alpha1.BootResource{}),
		elements:      kipxe.NewResources(),
	}
}

func (this *BootResources) Setup(logger logger.LogContext) {
	if this.initialized {
		return
	}
	this.initialized = true
	if logger != nil {
		logger.Infof("setup documents")
	}
	list, _ := this.resource.ListCached(labels.Everything())

	for _, l := range list {
		elem, err := this.Update(logger, l)
		if elem != nil {
			logger.Infof("found document %s", elem.Name())
		}
		if err != nil {
			logger.Infof("errorneous document %s: %s", l.GetName(), err)
		}
	}
}

func (this *BootResources) recheckUsers(logger logger.LogContext, users kipxe.NameSet) {
	logger.Infof("found users: %s", users)
	this.profiles.Recheck(users)
}

func (this *BootResources) Recheck(users kipxe.NameSet) {
	this.EnqueueAll(users, v1alpha1.RESOURCE)
	this.elements.Recheck(users)
}

func (this *BootResources) Update(logger logger.LogContext, obj resources.Object) (*kipxe.BootResource, error) {
	m, err := NewResource(obj, this.InfoBase.cache)
	if err == nil {
		this.recheckUsers(logger, this.elements.Set(m))
	}
	if err != nil {
		this.recheckUsers(logger, this.elements.Delete(obj.ObjectName()))
		logger.Errorf("invalid document: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*v1alpha1.BootResource)
			mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*v1alpha1.BootResource)
		mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_READY)
		mod.AssureStringValue(&m.Status.Message, "document ok")
		return nil
	})
	return m, err
}

func (this *BootResources) Delete(logger logger.LogContext, name resources.ObjectName) {
	this.recheckUsers(logger, this.elements.Delete(name))
}

func validateType(m *v1alpha1.BootResource) error {
	found := []string{}
	field := false
	if m.Spec.Text != "" {
		found = append(found, "text")
	}
	if m.Spec.Binary != "" {
		found = append(found, "binary")
	}
	if m.Spec.URL != "" {
		found = append(found, "url")
	}
	if m.Spec.ConfigMap != "" {
		field = true
		found = append(found, "configMap")
	}
	if m.Spec.Secret != "" {
		field = true
		found = append(found, "secret")
	}
	if len(found) > 1 {
		return fmt.Errorf("only one of %v can be used", found)
	}
	if m.Spec.FieldName != "" && !field && len(found) > 0 {
		return fmt.Errorf("field can only be used together with configMap, secret or metadata document")
	}
	return nil
}

func NewResource(obj resources.Object, cache kipxe.Cache) (*kipxe.BootResource, error) {
	var source kipxe.Source
	var err error

	m := obj.Data().(*v1alpha1.BootResource)
	mime := strings.TrimSpace(m.Spec.MimeType)
	if mime == "" {
		return nil, fmt.Errorf("mime type empty")
	}
	if err = validateType(m); err != nil {
		return nil, err
	}

	if m.Spec.Text != "" {
		_, err := template.New(m.Name).Parse(m.Spec.Text)
		if err != nil {
			return nil, fmt.Errorf("text is no valid go template: %s", err)
		}
		source = kipxe.NewTextSource(mime, m.Spec.Text)
	}

	if m.Spec.Binary != "" {
		s, err := kipxe.NewBinarySource(mime, m.Spec.Binary)
		if err != nil {
			return nil, fmt.Errorf("invalid binary data:%s", err)
		}
		source = s
	}

	if m.Spec.URL != "" {
		var src kipxe.URLSource
		if m.Spec.Volatile {
			cache = nil
		}
		src, err = kipxe.NewMappedURLSource(mime, m.Spec.URL, cache)
		if err != nil {
			return nil, err
		}
		if m.Spec.Redirect != nil && *m.Spec.Redirect {
			src = kipxe.NewURLRedirectSource(src)
		}
		source = src
	}

	if m.Spec.ConfigMap != "" {
		r, _ := obj.Resources().Get(&v1.ConfigMap{})
		source, err = NewMappedObjectSource(NewConfigMapSource(r, resources.NewObjectName(m.Namespace, m.Spec.ConfigMap), m.Spec.FieldName, mime))
	}
	if m.Spec.Secret != "" {
		r, _ := obj.Resources().Get(&v1.Secret{})
		source, err = NewMappedObjectSource(NewSecretSource(r, resources.NewObjectName(m.Namespace, m.Spec.Secret), m.Spec.FieldName, mime))
	}

	if err != nil {
		return nil, err
	}
	if source == nil {
		source = kipxe.NewMetaDataSource(mime, m.Spec.FieldName)
	}

	name := resources.NewObjectName(m.Namespace, m.Name)
	mapping, err := Mapping(fmt.Sprintf("resource %s(mapping)", name), m.Spec.Mapping)
	if err != nil {
		return nil, err
	}
	return kipxe.NewResource(name, mapping, m.Spec.Values.Values, source,
		m.Spec.Plain != nil && *m.Spec.Plain), nil
}

////////////////////////////////////////////////////////////////////////////////

type fieldFetcher func(obj runtime.Object, name string) ([]byte, error)

type objectSource struct {
	kipxe.SourceSupport
	resource resources.Interface
	name     resources.ObjectName
	field    string
	fetch    fieldFetcher
}

var _ kipxe.Source = &objectSource{}

func (this *objectSource) get() (resources.Object, error) {
	return this.resource.Get(this.name)
}

func (this *objectSource) Bytes() ([]byte, error) {
	obj, err := this.get()
	if err != nil {
		return nil, err
	}
	if this.field == "" {
		if this.MimeType() == kipxe.MIME_YAML {
			return yaml.Marshal(obj)
		}
		return json.Marshal(obj.Data())
	}
	return this.fetch(obj.Data(), this.field)
}

func (this *objectSource) Serve(w http.ResponseWriter, r *http.Request) {
	data, err := this.Bytes()
	if err != nil {
		if errors.IsNotFound(err) {
			this.IwriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("object %s not found", this.name))
		} else {
			this.IwriteErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("object %s: %s", this.name, err))
		}
		return
	}
	this.Iserve(data, w, r)
}

////////////////////////////////////////////////////////////////////////////////

type mappedObjectSource struct {
	*objectSource
	name  *template.Template
	field *template.Template
}

func NewMappedObjectSource(src *objectSource) (kipxe.Source, error) {
	name, err := kipxe.TemplateFor("name", src.name.Name())
	if err != nil {
		return nil, err
	}
	field, err := kipxe.TemplateFor("field", src.field)
	if err != nil {
		return nil, err
	}
	if name != nil || field != nil {
		return &mappedObjectSource{
			objectSource: src,
			name:         name,
			field:        field,
		}, nil
	}
	return src, nil
}

func (this *mappedObjectSource) Bytes() ([]byte, error) {
	return nil, fmt.Errorf("cannot serve object template")
}

func (this *mappedObjectSource) Serve(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write([]byte("cannot serve object template\n"))
}

func (this *mappedObjectSource) Map(values simple.Values) (kipxe.Source, error) {
	src := *this.objectSource
	if this.name != nil {
		buf := &strings.Builder{}
		err := this.name.Execute(buf, values)
		if err != nil {
			return nil, err
		}
		src.name = resources.NewObjectName(src.name.Namespace(), buf.String())
	}

	if this.field != nil {
		buf := &strings.Builder{}
		err := this.field.Execute(buf, values)
		if err != nil {
			return nil, err
		}
		src.field = buf.String()
	}
	return &src, nil
}

////////////////////////////////////////////////////////////////////////////////

func NewConfigMapSource(resc resources.Interface, name resources.ObjectName, field string, mimeType string) *objectSource {
	return &objectSource{
		SourceSupport: kipxe.NewSourceSupport(mimeType),
		resource:      resc,
		name:          name,
		field:         field,
		fetch: func(obj runtime.Object, field string) ([]byte, error) {
			cm := obj.(*v1.ConfigMap)
			data := cm.BinaryData[field]
			if data == nil {
				txt, ok := cm.Data[field]
				if !ok {
					return nil, fmt.Errorf("no field %s found", field)
				}
				data = []byte(txt)
			}
			return data, nil
		},
	}
}

func NewSecretSource(resc resources.Interface, name resources.ObjectName, field string, mimeType string) *objectSource {
	return &objectSource{
		SourceSupport: kipxe.NewSourceSupport(mimeType),
		resource:      resc,
		name:          name,
		field:         field,
		fetch: func(obj runtime.Object, field string) ([]byte, error) {
			cm := obj.(*v1.Secret)
			data := cm.Data[field]
			if data == nil {
				return nil, fmt.Errorf("no field %s found", field)
			}
			return data, nil
		},
	}
}
