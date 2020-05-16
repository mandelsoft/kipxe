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

package ready

import (
	"fmt"
	"sync"
)

type ReadyReporter interface {
	IsReady() bool
}

var lock sync.Mutex
var reporters []ReadyReporter

func Register(reporter ReadyReporter) {
	lock.Lock()
	defer lock.Unlock()

	reporters = append(reporters, reporter)
}

func ReadyInfo() (bool, string) {
	lock.Lock()
	defer lock.Unlock()

	if len(reporters) == 0 {
		return false, "no ready reported configured"
	}
	ready_cnt := 0
	notready_cnt := 0
	for _, r := range reporters {
		if r.IsReady() {
			ready_cnt++
		} else {
			notready_cnt++
		}
	}
	return notready_cnt == 0, fmt.Sprintf("ready: %d, not ready %d", ready_cnt, notready_cnt)
}
