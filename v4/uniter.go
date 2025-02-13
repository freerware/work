/* Copyright 2025 Freerware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package work

//Uniter represents a factory for work units.
type Uniter interface {

	//Unit constructs a new work unit.
	Unit() (Unit, error)
}

type uniter struct {
	options []UnitOption
}

// NewUniter creates a new uniter with the provided unit options.
func NewUniter(options ...UnitOption) Uniter {
	return uniter{options: options}
}

// Unit constructs a new work unit.
func (u uniter) Unit() (Unit, error) {
	return NewUnit(u.options...)
}
