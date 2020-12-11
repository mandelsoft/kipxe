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

package machines

import (
	"sync"
	"sync/atomic"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/cluster"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"
	"k8s.io/apimachinery/pkg/labels"

	api "github.com/onmetal/k8s-machines/pkg/apis/machines/v1alpha1"
)

/* REMARK:
 * when the metadata watches are implemented in cm lib a pure name indexer
 * makes sense to save memory
 * Getting the machine info then can be implmeneted by an api server get round trip.
 */

type FullIndexer struct {
	initlock    sync.RWMutex
	lock        sync.RWMutex
	initialized int32
	elements    map[resources.ObjectName]*Machine
	byMACs      map[string]*Machine
	byUUIDs     map[string]*Machine
}

func NewFullIndexer() *FullIndexer {
	m := &FullIndexer{
		elements: map[resources.ObjectName]*Machine{},
		byMACs:   map[string]*Machine{},
		byUUIDs:  map[string]*Machine{},
	}
	m.initlock.Lock()
	return m
}

func (this *FullIndexer) Wait() {
	this.initlock.RLock()
	this.initlock.RUnlock()
}

func (this *FullIndexer) GetByMAC(mac string) *Machine {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.byMACs[mac]
}

func (this *FullIndexer) GetByUUID(uuid string) *Machine {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.byUUIDs[uuid]
}

func (this *FullIndexer) GetByName(name resources.ObjectName) *Machine {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.elements[name]
}

func (this *FullIndexer) Setup(logger logger.LogContext, cluster cluster.Interface) error {
	if atomic.LoadInt32(&this.initialized) != 0 {
		logger.Infof("machine cache already initialized")
		return nil
	}
	if cluster == nil {
		logger.Infof("waiting for machine cache")
		this.Wait()
		return nil
	}

	resc, err := cluster.Resources().Get(api.MACHINEINFO)
	if err != nil {
		return err
	}
	logger.Infof("setup machines")
	list, _ := resc.ListCached(labels.Everything())

	for _, l := range list {
		elem, err, _ := ValidateMachine(logger, l)
		if elem != nil {
			this.Set(elem)
			logger.Infof("found machine %s", elem.Name)
		}
		if err != nil {
			logger.Infof("errorneous machine %s: %s", l.GetName(), err)
		}
	}
	logger.Infof("machine cache setup done")
	atomic.StoreInt32(&this.initialized, 1)
	this.initlock.Unlock()
	return nil
}

func (this *FullIndexer) Set(m *Machine) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	old := this.elements[m.Name]
	if old != nil {
		this.cleanup(old)
	}
	this.set(m)
	return nil
}

func (this *FullIndexer) Delete(name resources.ObjectName) {
	this.lock.Lock()
	defer this.lock.Unlock()
	old := this.elements[name]
	if old != nil {
		this.cleanup(old)
	}
}

func (this *FullIndexer) cleanup(m *Machine) {
	for _, n := range m.NICs {
		delete(this.byMACs, n.MAC)
	}
	delete(this.byUUIDs, m.UUID)
	delete(this.elements, m.Name)
}

func (this *FullIndexer) set(m *Machine) {
	for _, n := range m.NICs {
		this.byMACs[n.MAC] = m
	}
	if m.UUID != "" {
		this.byUUIDs[m.UUID] = m
	}
	this.elements[m.Name] = m
}
