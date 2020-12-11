/*
 * Copyright (c) 2020 by The metal-stack Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package v1alpha1

import (
	"github.com/gardener/controller-manager-library/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const STATE_OK = "Ok"
const STATE_INVALID = "Invalid"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineInfoList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineInfo `json:"items"`
}

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=machi,path=machineinfos,singular=machineinfo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name=UUID,JSONPath=".spec.uuid",type=string
// +kubebuilder:printcolumn:name=State,JSONPath=".status.state",type=string
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MachineInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MachineInfoSpec `json:"spec"`
	// +optional
	Status MachineInfoStatus `json:"status,omitempty"`
}

type MachineInfoSpec struct {
	// +optional
	UUID string `json:"uuid,omitempty"`
	// +optional
	NICs []NIC `json:"nics,omitempty"`
	// +optional
	CPUs []CPU `json:"cpus,omitempty"`
	// +optional
	Memory []Memory `json:"memory,omitempty"`
	// +optional
	Disks []Disk `json:"disks,omitempty"`

	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Values types.Values `json:"values,omitempty"`
}

type NIC struct {
	Name string `json:"name"`
	MAC  string `json:"mac"`
	// +optional
	Bandwidth int `json:"bandwidth,omitempty"`
}

type CPU struct {
	// +optional
	CPUInfo string `json:"cpuInfo,omitempty"`
	// +optional
	BogoMips int `json:"bogoMips,omitempty"`
	// +optional
	MHZ   int `json:"mhz,omitempty"`
	Cores int `json:"cores,omitempty"`
}

type Memory struct {
	Size int `json:"size"`
	// +kubebuilder:validation:XPreserveUnknownFields
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Numa types.Values `json:"numa,omitempty"`
}

type Disk struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

type MachineInfoStatus struct {
	// +optional
	State string `json:"state"`

	// +optional
	Message string `json:"message,omitempty"`
}
