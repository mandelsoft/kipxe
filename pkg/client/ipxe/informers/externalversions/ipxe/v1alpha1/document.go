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

// DocumentInformer provides access to a shared informer and lister for
// Documents.
type DocumentInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.DocumentLister
}

type documentInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewDocumentInformer constructs a new informer for BootResource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewDocumentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredDocumentInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredDocumentInformer constructs a new informer for BootResource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredDocumentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IpxeV1alpha1().Documents(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IpxeV1alpha1().Documents(namespace).Watch(options)
			},
		},
		&ipxev1alpha1.BootResource{},
		resyncPeriod,
		indexers,
	)
}

func (f *documentInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredDocumentInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *documentInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&ipxev1alpha1.BootResource{}, f.defaultInformer)
}

func (f *documentInformer) Lister() v1alpha1.DocumentLister {
	return v1alpha1.NewDocumentLister(f.Informer().GetIndexer())
}
