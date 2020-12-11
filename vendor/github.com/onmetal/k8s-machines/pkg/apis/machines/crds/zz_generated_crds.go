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
    controller-gen.kubebuilder.io/version: v0.2.9
  creationTimestamp: null
  name: machineinfos.machines.onmetal.de
spec:
  group: machines.onmetal.de
  names:
    kind: MachineInfo
    listKind: MachineInfoList
    plural: machineinfos
    shortNames:
    - machi
    singular: machineinfo
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
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              cpus:
                items:
                  properties:
                    bogoMips:
                      type: integer
                    cores:
                      type: integer
                    cpuInfo:
                      type: string
                    mhz:
                      type: integer
                  type: object
                type: array
              disks:
                items:
                  properties:
                    id:
                      type: string
                    name:
                      type: string
                    size:
                      type: integer
                    type:
                      type: string
                  required:
                  - id
                  - name
                  - size
                  - type
                  type: object
                type: array
              memory:
                items:
                  properties:
                    numa:
                      description: Values is used to specify an arbitrary document structure without the need of a regular manifest api group version as part of a kubernetes resource
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                    size:
                      type: integer
                  required:
                  - size
                  type: object
                type: array
              nics:
                items:
                  properties:
                    bandwidth:
                      type: integer
                    mac:
                      type: string
                    name:
                      type: string
                  required:
                  - mac
                  - name
                  type: object
                type: array
              uuid:
                type: string
              values:
                description: Values is used to specify an arbitrary document structure without the need of a regular manifest api group version as part of a kubernetes resource
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
}

func AddToRegistry(r apiextensions.Registry) {
	registry.AddToRegistry(r)
}
