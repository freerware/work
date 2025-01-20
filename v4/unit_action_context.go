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

import (
	"github.com/uber-go/tally/v4"
)

// UnitActionContext represents the executional context for an action.
type UnitActionContext struct {
	// Logger is the work units configured logger.
	Logger UnitLogger
	// Scope is the work units configured metrics scope.
	Scope tally.Scope
	// AdditionCount represents the number of entities indicated as new.
	AdditionCount int
	// AlterationCount represents the number of entities indicated as modified.
	AlterationCount int
	// RemovalCount represents the number of entities indicated as removed.
	RemovalCount int
	// RegisterCount represents the number of entities indicated as registered.
	RegisterCount int
}
