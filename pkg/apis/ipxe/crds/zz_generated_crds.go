/*
Copyright (c) YEAR SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

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

package crds

import (
	"github.com/gardener/controller-manager-library/pkg/resources/apiextensions"
	"github.com/gardener/controller-manager-library/pkg/utils"
)

var registry = apiextensions.NewRegistry()

func init() {
	var data string
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: bootprofilematchers.ipxe.mandelsoft.org
spec:
  group: ipxe.mandelsoft.org
  names:
    kind: BootProfileMatcher
    listKind: BootProfileMatcherList
    plural: bootprofilematchers
    shortNames:
    - bmatch
    singular: bootprofilematcher
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.profileName
      name: Profile
      type: string
    - jsonPath: .status.state
      name: State
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              mapping:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
              matcher:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
                x-kubernetes-preserve-unknown-fields: true
              profileName:
                type: string
              selector:
                description: A label selector is a label query over a set of resources.
                  The result of matchLabels and matchExpressions are ANDed. An empty
                  label selector matches all objects. A null label selector matches
                  no objects.
                properties:
                  matchExpressions:
                    description: matchExpressions is a list of label selector requirements.
                      The requirements are ANDed.
                    items:
                      description: A label selector requirement is a selector that
                        contains values, a key, and an operator that relates the key
                        and values.
                      properties:
                        key:
                          description: key is the label key that the selector applies
                            to.
                          type: string
                        operator:
                          description: operator represents a key's relationship to
                            a set of values. Valid operators are In, NotIn, Exists
                            and DoesNotExist.
                          type: string
                        values:
                          description: values is an array of string values. If the
                            operator is In or NotIn, the values array must be non-empty.
                            If the operator is Exists or DoesNotExist, the values
                            array must be empty. This array is replaced during a strategic
                            merge patch.
                          items:
                            type: string
                          type: array
                      required:
                      - key
                      - operator
                      type: object
                    type: array
                  matchLabels:
                    additionalProperties:
                      type: string
                    description: matchLabels is a map of {key,value} pairs. A single
                      {key,value} in the matchLabels map is equivalent to an element
                      of matchExpressions, whose key field is "key", the operator
                      is "In", and the values array contains only "value". The requirements
                      are ANDed.
                    type: object
                type: object
              values:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
              weight:
                type: integer
            required:
            - profileName
            type: object
          status:
            properties:
              message:
                type: string
              state:
                type: string
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	utils.Must(registry.RegisterCRD(data))
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: bootprofiles.ipxe.mandelsoft.org
spec:
  group: ipxe.mandelsoft.org
  names:
    kind: BootProfile
    listKind: BootProfileList
    plural: bootprofiles
    shortNames:
    - bprof
    singular: bootprofile
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .status.state
      name: State
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              mapping:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
                x-kubernetes-preserve-unknown-fields: true
              resources:
                items:
                  properties:
                    documentName:
                      type: string
                    path:
                      type: string
                  required:
                  - documentName
                  - path
                  type: object
                type: array
              values:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
                x-kubernetes-preserve-unknown-fields: true
            type: object
          status:
            properties:
              message:
                type: string
              state:
                type: string
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	utils.Must(registry.RegisterCRD(data))
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: bootresources.ipxe.mandelsoft.org
spec:
  group: ipxe.mandelsoft.org
  names:
    kind: BootResource
    listKind: BootResourceList
    plural: bootresources
    shortNames:
    - bresc
    singular: bootresource
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .status.state
      name: State
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              URL:
                type: string
              binary:
                type: string
              mapping:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
                x-kubernetes-preserve-unknown-fields: true
              mimeType:
                type: string
              plainContent:
                type: boolean
              redirect:
                type: boolean
              text:
                type: string
              values:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
              volatile:
                type: boolean
            type: object
          status:
            properties:
              message:
                type: string
              state:
                type: string
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	utils.Must(registry.RegisterCRD(data))
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: machines.ipxe.mandelsoft.org
spec:
  group: ipxe.mandelsoft.org
  names:
    kind: Machine
    listKind: MachineList
    plural: machines
    shortNames:
    - mach
    singular: machine
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.uuid
      name: UUID
      type: string
    - jsonPath: .status.state
      name: State
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              additional:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
                x-kubernetes-preserve-unknown-fields: true
              macs:
                additionalProperties:
                  items:
                    type: string
                  type: array
                type: object
                x-kubernetes-preserve-unknown-fields: true
              uuid:
                type: string
              values:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
                x-kubernetes-preserve-unknown-fields: true
            type: object
          status:
            properties:
              message:
                type: string
              state:
                type: string
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	utils.Must(registry.RegisterCRD(data))
	data = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.4
  creationTimestamp: null
  name: metadatamappers.ipxe.mandelsoft.org
spec:
  group: ipxe.mandelsoft.org
  names:
    kind: MetaDataMapper
    listKind: MetaDataMapperList
    plural: metadatamappers
    shortNames:
    - mdmap
    singular: metadatamapper
  scope: Namespaced
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.weight
      name: Weight
      type: integer
    - jsonPath: .spec.URL
      name: URL
      type: string
    - jsonPath: .status.state
      name: State
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              URL:
                type: string
              mapping:
                description: Values is a workarround for kubebuilder to be able to
                  generate an API spec. The Values MUST be marked with "-" to avoud
                  errors.
                type: object
              weight:
                type: integer
            required:
            - weight
            type: object
          status:
            properties:
              message:
                type: string
              state:
                type: string
            type: object
        required:
        - spec
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
  `
	utils.Must(registry.RegisterCRD(data))
}

func AddToRegistry(r apiextensions.Registry) {
	registry.AddToRegistry(r)
}
