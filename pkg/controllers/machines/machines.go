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

package machines

import (
	"net/http"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/mandelsoft/kipxe/pkg/apis/ipxe/v1alpha1"
	"github.com/mandelsoft/kipxe/pkg/controllers"
	"github.com/mandelsoft/kipxe/pkg/kipxe"
)

type Machine struct {
	name   resources.ObjectName
	uuid   string
	macs   v1alpha1.MachineMACs
	values simple.Values
}

type Machines struct {
	ResourceCache
	lock     sync.RWMutex
	elements map[resources.ObjectName]*Machine
	byMACs   map[string]*Machine
	byUUIDs  map[string]*Machine
}

func newMachines(controller controller.Interface) *Machines {
	m := &Machines{
		ResourceCache: NewResourceCache(controller, &v1alpha1.Machine{}),
		elements:      map[resources.ObjectName]*Machine{},
		byMACs:        map[string]*Machine{},
		byUUIDs:       map[string]*Machine{},
	}
	controller.Infof("registering machine metadata mapping")
	controllers.GetSharedRegistry(controller).Register(m)
	return m
}

func (this *Machines) Setup(logger logger.LogContext) {
	if this.initialized {
		return
	}
	this.initialized = true
	if logger != nil {
		logger.Infof("setup machines")
	}
	list, _ := this.resource.ListCached(labels.Everything())

	for _, l := range list {
		elem, err := this.Update(logger, l)
		if elem != nil {
			logger.Infof("found machine %s", elem.name)
		}
		if err != nil {
			logger.Infof("errorneous machine %s: %s", l.GetName(), err)
		}
	}
}

func (this *Machines) Set(m *Machine) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	old := this.elements[m.name]
	if old != nil {
		this.cleanup(old)
	}
	this.set(m)
	return nil
}

func (this *Machines) Delete(logger logger.LogContext, name resources.ObjectName) {
	this.lock.Lock()
	defer this.lock.Unlock()
	old := this.elements[name]
	if old != nil {
		this.cleanup(old)
	}
}

func (this *Machines) cleanup(m *Machine) {
	for _, l := range m.macs {
		for _, mac := range l {
			delete(this.byMACs, mac)
		}
	}
	delete(this.byUUIDs, m.uuid)
	delete(this.elements, m.name)
}

func (this *Machines) set(m *Machine) {
	for _, l := range m.macs {
		for _, mac := range l {
			this.byMACs[mac] = m
		}
	}
	if m.uuid != "" {
		this.byUUIDs[m.uuid] = m
	}
	this.elements[m.name] = m
}

func (this *Machines) Lookup(values kipxe.MetaData) *Machine {
	this.lock.RLock()
	defer this.lock.RUnlock()

	uuid := values["uuid"]
	if uuid != nil {
		m := this.byUUIDs[uuid.(string)]
		if m != nil {
			return m
		}
	}
	macs := values["__mac__"]
	if macs != nil {
		for _, v := range macs.([]interface{}) {
			m := this.byMACs[v.(string)]
			if m != nil {
				return m
			}
		}
	}
	return nil
}

func (this *Machines) Update(logger logger.LogContext, obj resources.Object) (*Machine, error) {
	m, err := NewMachine(obj.Data().(*v1alpha1.Machine))
	if err == nil {
		err = this.Set(m)
	}
	if err != nil {
		logger.Errorf("invalid machine: %s", err)
		_, err2 := resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
			m := mod.Data().(*v1alpha1.Machine)
			mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_INVALID)
			mod.AssureStringValue(&m.Status.Message, err.Error())
			return nil
		})
		return nil, err2
	}
	_, err = resources.ModifyStatus(obj, func(mod *resources.ModificationState) error {
		m := mod.Data().(*v1alpha1.Machine)
		mod.AssureStringValue(&m.Status.State, v1alpha1.STATE_READY)
		mod.AssureStringValue(&m.Status.Message, "machine ok")
		return nil
	})
	return m, err
}

func (this *Machines) Weight() int {
	return 100
}

func (this *Machines) String() string {
	return "machine controller"
}

func (this *Machines) Map(logger logger.LogContext, values kipxe.MetaData, req *http.Request) (kipxe.MetaData, error) {
	values = values.DeepCopy()
	m := this.Lookup(values)
	if m != nil {
		logger.Infof("found machine %s", m.name)
		if m.uuid != "" {
			values["uuid"] = m.uuid
		}
		values["attributes"] = v1alpha1.CopyAndNormalize(m.values)
		values["macsbypurpose"] = v1alpha1.CopyAndNormalize(m.macs)
	} else {
		logger.Infof("no machine found -> trigger registration")
		values["task"] = "register"
	}
	return values, nil
}

func NewMachine(m *v1alpha1.Machine) (*Machine, error) {
	values := m.Spec.Values.Values
	macs := m.Spec.MACs

	if values == nil {
		values = simple.Values{}
	}
	if macs == nil {
		macs = v1alpha1.MachineMACs{}
	}
	return &Machine{
		name:   resources.NewObjectName(m.Namespace, m.Name),
		uuid:   m.Spec.UUID,
		macs:   macs,
		values: values,
	}, nil
}
