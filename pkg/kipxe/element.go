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

type ElementChecker func() error

type Element struct {
	Named
	error error
}

func NewElement(name Name) Element {
	return Element{
		Named: NewNamed(name),
		error: nil,
	}
}

func (this *Element) Error() error {
	return this.error
}

func (this *Element) recheck(checker ElementChecker) bool {
	if checker == nil {
		return false
	}
	new := checker()
	if new == this.error {
		return false
	}

	defer func() { this.error = new }()
	if new != nil && this.error != nil {
		if new.Error() == this.error.Error() {
			return false
		}
	}
	return true
}
