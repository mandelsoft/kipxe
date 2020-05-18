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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BootProfileMatcherList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BootProfileMatcher `json:"items"`
}

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=bmatch,path=bootprofilematchers,singular=bootprofilematcher
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=Profile,JSONPath=".spec.profileName",type=string
// +kubebuilder:printcolumn:name=State,JSONPath=".status.state",type=string
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BootProfileMatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BootProfileMatcherSpec `json:"spec"`
	// +optional
	Status BootProfileMatcherStatus `json:"status,omitempty"`
}

type BootProfileMatcherSpec struct {
	// +optional
	Selector *metav1.LabelSelector `json:"selector"`
	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Mapping Values `json:"mapping,omitempty"`
	// +optional
	Values Values `json:"values,omitempty"`
	// +optional
	Weight  *int   `json:"weight,omitempty"`
	Profile string `json:"profileName"`
}

type BootProfileMatcherStatus struct {
	// +optional
	State string `json:"state"`

	// +optional
	Message string `json:"message,omitempty"`
}
