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

package kipxe

import (
	"sync"
)

const EVT_INFO = "info"
const EVT_WARN = "warn"
const EVT_ERR = "error"

const EVT_MAPPER = "mapper"
const EVT_MATCHER = "matcher"
const EVT_PROFILE = "profile"
const EVT_RESOURCE = "resource"

type EventHandler interface {
	HandleEvent(name Name, otype, etype string, msg string, args ...interface{})
}

type EventHandlers struct {
	lock     sync.RWMutex
	handlers []EventHandler
}

func (this *EventHandlers) Register(h EventHandler) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, e := range this.handlers {
		if e == h {
			return
		}
	}
	this.handlers = append(this.handlers, h)
}

func (this *EventHandlers) UnRegister(h EventHandler) {
	this.lock.Lock()
	defer this.lock.Unlock()
	for i, e := range this.handlers {
		if e == h {
			this.handlers = append(this.handlers[:i], this.handlers[i+1:]...)
		}
	}
}

func (this *EventHandlers) HandleEvent(name Name, otype, etype, msg string, args ...interface{}) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	for _, e := range this.handlers {
		e.HandleEvent(name, otype, etype, msg, args)
	}
}
