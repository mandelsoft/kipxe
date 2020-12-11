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

package controllers

import (
	"sync"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"

	"github.com/onmetal/k8s-machines/pkg/machines"
)

var key = ctxutil.SimpleKey("machineindex")

func GetOrCreateMachineIndex(env extension.Environment) machines.MachineIndexer {
	return env.ControllerManager().GetOrCreateSharedValue(key, func() interface{} {
		return machines.NewFullIndexer()
	}).(machines.MachineIndexer)
}

func GetMachineIndex(env extension.Environment) machines.MachineIndexer {
	i := env.ControllerManager().GetSharedValue(key)
	if i != nil {
		return i.(machines.MachineIndexer)
	}
	return nil
}

type Client interface {
	PropagateMachineIndex(machines.MachineIndex)
}

var registry []Client
var lock sync.Mutex

func RegisterClient(c Client) {
	lock.Lock()
	defer lock.Unlock()
	registry = append(registry, c)
}

func PropagateMachineInfos(index machines.MachineIndex) {
	lock.Lock()
	defer lock.Unlock()
	for _, c := range registry {
		c.PropagateMachineIndex(index)
	}
}
