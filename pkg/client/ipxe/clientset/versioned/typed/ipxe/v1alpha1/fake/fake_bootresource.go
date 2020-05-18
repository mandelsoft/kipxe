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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBootResources implements BootResourceInterface
type FakeBootResources struct {
	Fake *FakeIpxeV1alpha1
	ns   string
}

var bootresourcesResource = schema.GroupVersionResource{Group: "ipxe.mandelsoft.org", Version: "v1alpha1", Resource: "bootresources"}

var bootresourcesKind = schema.GroupVersionKind{Group: "ipxe.mandelsoft.org", Version: "v1alpha1", Kind: "BootResource"}

// Get takes name of the bootResource, and returns the corresponding bootResource object, and an error if there is any.
func (c *FakeBootResources) Get(name string, options v1.GetOptions) (result *v1alpha1.BootResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(bootresourcesResource, c.ns, name), &v1alpha1.BootResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BootResource), err
}

// List takes label and field selectors, and returns the list of BootResources that match those selectors.
func (c *FakeBootResources) List(opts v1.ListOptions) (result *v1alpha1.BootResourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(bootresourcesResource, bootresourcesKind, c.ns, opts), &v1alpha1.BootResourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BootResourceList{ListMeta: obj.(*v1alpha1.BootResourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.BootResourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bootResources.
func (c *FakeBootResources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(bootresourcesResource, c.ns, opts))

}

// Create takes the representation of a bootResource and creates it.  Returns the server's representation of the bootResource, and an error, if there is any.
func (c *FakeBootResources) Create(bootResource *v1alpha1.BootResource) (result *v1alpha1.BootResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(bootresourcesResource, c.ns, bootResource), &v1alpha1.BootResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BootResource), err
}

// Update takes the representation of a bootResource and updates it. Returns the server's representation of the bootResource, and an error, if there is any.
func (c *FakeBootResources) Update(bootResource *v1alpha1.BootResource) (result *v1alpha1.BootResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(bootresourcesResource, c.ns, bootResource), &v1alpha1.BootResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BootResource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeBootResources) UpdateStatus(bootResource *v1alpha1.BootResource) (*v1alpha1.BootResource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(bootresourcesResource, "status", c.ns, bootResource), &v1alpha1.BootResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BootResource), err
}

// Delete takes name of the bootResource and deletes it. Returns an error if one occurs.
func (c *FakeBootResources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(bootresourcesResource, c.ns, name), &v1alpha1.BootResource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBootResources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(bootresourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BootResourceList{})
	return err
}

// Patch applies the patch and returns the patched bootResource.
func (c *FakeBootResources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BootResource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(bootresourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.BootResource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BootResource), err
}