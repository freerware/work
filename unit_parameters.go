/* Copyright 2019 Freerware
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
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// UnitParameters represents the collection of
// dependencies and configuration needed for a work unit.
type UnitParameters struct {

	//Inserters indicates the mappings between inserters
	//and the entity types they insert.
	Inserters map[TypeName]Inserter

	//Updates indicates the mappings between updaters
	//and the entity types they update.
	Updaters map[TypeName]Updater

	//Deleters indicates the mappings between deleters
	//and the entity types they delete.
	Deleters map[TypeName]Deleter

	//Logger represents the logger that the work unit will utilize.
	Logger *zap.Logger

	//Scope represents the metric scope that the work unit will utilize.
	Scope tally.Scope
}
