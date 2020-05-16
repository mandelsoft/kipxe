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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	ipxev1alpha1 "github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	versioned "github.com/mandelsoft/kipxe/pkg/client/ipxe/clientset/versioned"
	internalinterfaces "github.com/mandelsoft/kipxe/pkg/client/ipxe/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/mandelsoft/kipxe/pkg/client/ipxe/listers/ipxe/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MatcherInformer provides access to a shared informer and lister for
// Matchers.
type MatcherInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MatcherLister
}

type matcherInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMatcherInformer constructs a new informer for Matcher type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMatcherInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMatcherInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMatcherInformer constructs a new informer for Matcher type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMatcherInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IpxeV1alpha1().Matchers(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IpxeV1alpha1().Matchers(namespace).Watch(options)
			},
		},
		&ipxev1alpha1.Matcher{},
		resyncPeriod,
		indexers,
	)
}

func (f *matcherInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMatcherInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *matcherInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&ipxev1alpha1.Matcher{}, f.defaultInformer)
}

func (f *matcherInformer) Lister() v1alpha1.MatcherLister {
	return v1alpha1.NewMatcherLister(f.Informer().GetIndexer())
}
