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

import "fmt"

// TypeName represents an entity's type.
type TypeName string

// TypeNameOf provides the type name for the provided entity.
func TypeNameOf(entity interface{}) TypeName {
	return TypeName(fmt.Sprintf("%T", entity))
}

// String provides the string representation of the type name.
func (t TypeName) String() string {
	return string(t)
}
