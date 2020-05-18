/*
Copyright (c) 2020 Mandelsoft. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BootProfileMatcherLister helps list BootProfileMatchers.
type BootProfileMatcherLister interface {
	// List lists all BootProfileMatchers in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.BootProfileMatcher, err error)
	// BootProfileMatchers returns an object that can list and get BootProfileMatchers.
	BootProfileMatchers(namespace string) BootProfileMatcherNamespaceLister
	BootProfileMatcherListerExpansion
}

// bootProfileMatcherLister implements the BootProfileMatcherLister interface.
type bootProfileMatcherLister struct {
	indexer cache.Indexer
}

// NewBootProfileMatcherLister returns a new BootProfileMatcherLister.
func NewBootProfileMatcherLister(indexer cache.Indexer) BootProfileMatcherLister {
	return &bootProfileMatcherLister{indexer: indexer}
}

// List lists all BootProfileMatchers in the indexer.
func (s *bootProfileMatcherLister) List(selector labels.Selector) (ret []*v1alpha1.BootProfileMatcher, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BootProfileMatcher))
	})
	return ret, err
}

// BootProfileMatchers returns an object that can list and get BootProfileMatchers.
func (s *bootProfileMatcherLister) BootProfileMatchers(namespace string) BootProfileMatcherNamespaceLister {
	return bootProfileMatcherNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// BootProfileMatcherNamespaceLister helps list and get BootProfileMatchers.
type BootProfileMatcherNamespaceLister interface {
	// List lists all BootProfileMatchers in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.BootProfileMatcher, err error)
	// Get retrieves the BootProfileMatcher from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.BootProfileMatcher, error)
	BootProfileMatcherNamespaceListerExpansion
}

// bootProfileMatcherNamespaceLister implements the BootProfileMatcherNamespaceLister
// interface.
type bootProfileMatcherNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all BootProfileMatchers in the indexer for a given namespace.
func (s bootProfileMatcherNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.BootProfileMatcher, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BootProfileMatcher))
	})
	return ret, err
}

// Get retrieves the BootProfileMatcher from the indexer for a given namespace and name.
func (s bootProfileMatcherNamespaceLister) Get(name string) (*v1alpha1.BootProfileMatcher, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("bootprofilematcher"), name)
	}
	return obj.(*v1alpha1.BootProfileMatcher), nil
}
