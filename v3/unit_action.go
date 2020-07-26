/* Copyright 2020 Freerware
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

// Action represents an operation performed during a paticular lifecycle event of a work unit.
type UnitAction func(UnitActionContext)

// UnitActionType represents the type of work unit action.
type UnitActionType int

// The various types of actions that are executed throughout the lifecycle of a work unit.
const (
	// UnitActionTypeAfterRegister indicates an action type that occurs after an entity is registered.
	UnitActionTypeAfterRegister = iota
	// UnitActionTypeAfterAdd indicates an action type that occurs after an entity is added.
	UnitActionTypeAfterAdd
	// UnitActionTypeAfterAlter indicates an action type that occurs after an entity is altered.
	UnitActionTypeAfterAlter
	// UnitActionTypeAfterRemove indicates an action type that occurs after an entity is removed.
	UnitActionTypeAfterRemove
	// UnitActionTypeAfterInserts indicates an action type that occurs after new entities are inserted in the data store.
	UnitActionTypeAfterInserts
	// UnitActionTypeAfterUpdates indicates an action type that occurs after existing entities are updated in the data store.
	UnitActionTypeAfterUpdates
	// UnitActionTypeAfterDeletes indicates an action type that occurs after existing entities are deleted in the data store.
	UnitActionTypeAfterDeletes
	// UnitActionTypeAfterRollback indicates an action type that occurs after rollback.
	UnitActionTypeAfterRollback
	// UnitActionTypeAfterSave indicates an action type that occurs after save.
	UnitActionTypeAfterSave
	// UnitActionTypeBeforeRegister indicates an action type that occurs before an entity is registered.
	UnitActionTypeBeforeRegister
	// UnitActionTypeBeforeAdd indicates an action type that occurs before an entity is added.
	UnitActionTypeBeforeAdd
	// UnitActionTypeBeforeAlter indicates an action type that occurs before an entity is altered.
	UnitActionTypeBeforeAlter
	// UnitActionTypeBeforeRemove indicates an action type that occurs before an entity is removed.
	UnitActionTypeBeforeRemove
	// UnitActionTypeBeforeInserts indicates an action type that occurs before new entities are inserted in the data store.
	UnitActionTypeBeforeInserts
	// UnitActionTypeBeforeUpdates indicates an action type that occurs before existing entities are updated in the data store.
	UnitActionTypeBeforeUpdates
	// UnitActionTypeBeforeDeletes indicates an action type that occurs before existing entities are deleted in the data store.
	UnitActionTypeBeforeDeletes
	// UnitActionTypeBeforeRollback indicates an action type that occurs before rollback.
	UnitActionTypeBeforeRollback
	// UnitActionTypeBeforeSave indicates an action type that occurs before save.
	UnitActionTypeBeforeSave
)
