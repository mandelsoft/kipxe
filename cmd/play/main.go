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

package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	values "github.com/gardener/controller-manager-library/pkg/types/infodata/simple"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type bla string

func main() {
	v := values.Values{
		"v": map[string]interface{}{
			"x": 1,
			"z": 32.2,
		},
		"a": []interface{}{
			"v1",
		},
	}

	V := v1alpha1.Values{kipxe.NormValues(v)}
	s, _ := json.MarshalIndent(V.DeepCopy(), "", "  ")

	fmt.Printf("%s\n", s)

	u := &v1alpha1.Values{}
	err := json.Unmarshal(s, u)
	if err != nil {
		panic(err)
	}
	fmt.Printf("=== %v\n", u)

	var i interface{}

	m := reflect.TypeOf(map[string]interface{}{})
	value := reflect.ValueOf(v)
	if value.Type().ConvertibleTo(m) {
		i = value.Convert(m).Interface()
	}
	switch i.(type) {
	case map[string]interface{}:
		fmt.Printf("found map\n")
	default:
		fmt.Printf("found something else\n")
	}

	t, err := template.New("{.url}").Parse("{{.url}}")
	if err != nil {
		panic(err)
	}
	buf := &strings.Builder{}
	err = t.Execute(buf, map[string]interface{}{
		"url": "https://google.de",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("URL: %q\n", buf.String())

	md := kipxe.MetaData{}
	md["key"] = "value"
	md["foo"] = map[string]interface{}{
		"bar": 5,
	}

	fmt.Printf("false: %t, %t, %t\n", md.Has("key1"), md.Has("foo/x"), md.Has("foo/bar/x"))
	fmt.Printf("true : %t, %t, %t\n", md.Has("key"), md.Has("foo"), md.Has("foo/bar"))
	fmt.Printf("val  : %v, %v, %v\n", md.Get("key"), md.Get("foo"), md.Get("foo/bar"))

	I := v1alpha1.Values{}
	_ = map[string]interface{}(I.Values)

	x := bla("test")

	fmt.Printf("x: %s\n", kipxe.AsString(&x))
	fmt.Printf("map: %s\n", kipxe.AsString(map[string]interface{}{"a": "b"}))
}
