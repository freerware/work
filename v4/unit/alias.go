/* Copyright 2022 Freerware
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

package unit

import (
	"github.com/freerware/work/v4"
)

/* Errors. */

var (
	// ErrMissingDataMapper represents the error that is returned
	// when attempting to add, alter, remove, or register an entity
	// that doesn't have a corresponding data mapper.
	ErrMissingDataMapper = work.ErrMissingDataMapper
	// ErrNoDataMapper represents the error that occurs when attempting
	// to create a work unit without any data mappers.
	ErrNoDataMapper = work.ErrNoDataMapper
)

/* Units + Uniters. */

// Unit represents an atomic set of entity changes.
type Unit = work.Unit

// Uniter represents a factory for work units.
type Uniter = work.Uniter

// TypeName represents an entity's type.
type TypeName = work.TypeName

var (
	// TypeNameOf provides the type name for the provided entity.
	TypeNameOf = work.TypeNameOf
	// New creates a new work unit.
	New = work.NewUnit
	// NewUniter creates a new uniter with the provided unit options.
	NewUniter = work.NewUniter
)

/* Options. */

// Option applies an option to the provided configuration.
type Option = work.UnitOption

// Options represents the configuration options for the work unit.
type Options = work.UnitOptions

// RetryDelayType represents the type of retry delay to perform.
type RetryDelayType = work.UnitRetryDelayType

var (
	// DB specifies the option to provide the database for the work unit.
	DB = work.UnitDB
	// DataMappers specifies the option to provide the data mappers for
	// the work unit.
	DataMappers = work.UnitDataMappers
	// Logger specifies the option to provide a logger for the work unit.
	Logger = work.UnitLogger
	// Scope specifies the option to provide a metric scope for the work unit.
	Scope = work.UnitScope
	// AfterRegisterActions specifies the option to provide actions to execute
	// after entities are registered with the work unit.
	AfterRegisterActions = work.UnitAfterRegisterActions
	// AfterAddActions specifies the option to provide actions to execute
	// after entities are added with the work unit.
	AfterAddActions = work.UnitAfterAddActions
	// AfterAlterActions specifies the option to provide actions to execute
	// after entities are altered with the work unit.
	AfterAlterActions = work.UnitAfterAlterActions
	// AfterRemoveActions specifies the option to provide actions to execute
	// after entities are removed with the work unit.
	AfterRemoveActions = work.UnitAfterRemoveActions
	// AfterInsertsActions specifies the option to provide actions to execute
	// after new entities are inserted in the data store.
	AfterInsertsActions = work.UnitAfterInsertsActions
	// AfterUpdatesActions specifies the option to provide actions to execute
	// after altered entities are updated in the data store.
	AfterUpdatesActions = work.UnitAfterUpdatesActions
	// AfterDeletesActions specifies the option to provide actions to execute
	// after removed entities are deleted in the data store.
	AfterDeletesActions = work.UnitAfterDeletesActions
	// AfterRollbackActions specifies the option to provide actions to execute
	// after a rollback is performed.
	AfterRollbackActions = work.UnitAfterRollbackActions
	// AfterSaveActions specifies the option to provide actions to execute
	// after a save is performed.
	AfterSaveActions = work.UnitAfterSaveActions
	// BeforeInsertsActions specifies the option to provide actions to execute
	// before new entities are inserted in the data store.
	BeforeInsertsActions = work.UnitBeforeInsertsActions
	// BeforeUpdatesActions specifies the option to provide actions to execute
	// before altered entities are updated in the data store.
	BeforeUpdatesActions = work.UnitBeforeUpdatesActions
	// BeforeDeletesActions specifies the option to provide actions to execute
	// before removed entities are deleted in the data store.
	BeforeDeletesActions = work.UnitBeforeDeletesActions
	// BeforeRollbackActions specifies the option to provide actions to execute
	// before a rollback is performed.
	BeforeRollbackActions = work.UnitBeforeRollbackActions
	// BeforeSaveActions specifies the option to provide actions to execute
	// before a save is performed.
	BeforeSaveActions = work.UnitBeforeSaveActions
	// DefaultLoggingActions specifies all of the default logging actions.
	DefaultLoggingActions = work.UnitDefaultLoggingActions
	// DisableDefaultLoggingActions disables the default logging actions.
	DisableDefaultLoggingActions = work.DisableDefaultLoggingActions
	// RetryAttempts defines the number of retry attempts to perform.
	RetryAttempts = work.UnitRetryAttempts
	// RetryDelay defines the delay to utilize during retries.
	RetryDelay = work.UnitRetryDelay
	// RetryMaximumJitter defines the maximum jitter to utilize during
	// retries that utilize random delay times.
	RetryMaximumJitter = work.UnitRetryMaximumJitter
	// RetryType defines the type of retry to perform.
	RetryType = work.UnitRetryType
	// InsertFunc defines the function to be used for inserting new
	// entities in the underlying data store.
	InsertFunc = work.UnitInsertFunc
	// UpdateFunc defines the function to be used for updating existing
	// entities in the underlying data store.
	UpdateFunc = work.UnitUpdateFunc
	// DeleteFunc defines the function to be used for deleting existing
	// entities in the underlying data store.
	DeleteFunc = work.UnitDeleteFunc
)

/* Actions. */

// ActionContext represents the executional context for an action.
type ActionContext = work.UnitActionContext

// Action represents an operation performed during a paticular lifecycle
// event of a work unit.
type Action = work.UnitAction

// ActionType represents the type of work unit action.
type ActionType = work.UnitActionType

var (
	// ActionTypeAfterRegister indicates an action type that occurs after
	// an entity is registered.
	ActionTypeAfterRegister = work.UnitActionTypeAfterRegister
	// ActionTypeAfterAdd indicates an action type that occurs after an
	// entity is added.
	ActionTypeAfterAdd = work.UnitActionTypeAfterAdd
	// ActionTypeAfterAlter indicates an action type that occurs after
	// an entity is altered.
	ActionTypeAfterAlter = work.UnitActionTypeAfterAlter
	// ActionTypeAfterRemove indicates an action type that occurs after
	// an entity is removed.
	ActionTypeAfterRemove = work.UnitActionTypeAfterRemove
	// ActionTypeAfterInserts indicates an action type that occurs after
	// new entities are inserted in the data store.
	ActionTypeAfterInserts = work.UnitActionTypeAfterInserts
	// ActionTypeAfterUpdates indicates an action type that occurs after
	// existing entities are updated in the data store.
	ActionTypeAfterUpdates = work.UnitActionTypeAfterUpdates
	// ActionTypeAfterDeletes indicates an action type that occurs after
	// existing entities are deleted in the data store.
	ActionTypeAfterDeletes = work.UnitActionTypeAfterDeletes
	// ActionTypeAfterRollback indicates an action type that occurs after
	// rollback.
	ActionTypeAfterRollback = work.UnitActionTypeAfterRollback
	// ActionTypeAfterSave indicates an action type that occurs after save.
	ActionTypeAfterSave = work.UnitActionTypeAfterSave
	// ActionTypeBeforeRegister indicates an action type that occurs
	// before an entity is registered.
	ActionTypeBeforeRegister = work.UnitActionTypeBeforeRegister
	// ActionTypeBeforeAdd indicates an action type that occurs before an
	// entity is added.
	ActionTypeBeforeAdd = work.UnitActionTypeBeforeAdd
	// ActionTypeBeforeAlter indicates an action type that occurs before an
	// entity is altered.
	ActionTypeBeforeAlter = work.UnitActionTypeBeforeAlter
	// ActionTypeBeforeRemove indicates an action type that occurs before an
	// entity is removed.
	ActionTypeBeforeRemove = work.UnitActionTypeBeforeRemove
	// ActionTypeBeforeInserts indicates an action type that occurs before
	// new entities are inserted in the data store.
	ActionTypeBeforeInserts = work.UnitActionTypeBeforeInserts
	// ActionTypeBeforeUpdates indicates an action type that occurs before
	// existing entities are updated in the data store.
	ActionTypeBeforeUpdates = work.UnitActionTypeBeforeUpdates
	// ActionTypeBeforeDeletes indicates an action type that occurs before
	// existing entities are deleted in the data store.
	ActionTypeBeforeDeletes = work.UnitActionTypeBeforeDeletes
	// ActionTypeBeforeRollback indicates an action type that occurs before
	// rollback.
	ActionTypeBeforeRollback = work.UnitActionTypeBeforeRollback
	// ActionTypeBeforeSave indicates an action type that occurs before save.
	ActionTypeBeforeSave = work.UnitActionTypeBeforeSave
)

/* Data Mappers. */

// MapperContext represents the executional context for a data mapper.
type MapperContext = work.MapperContext

// DataMapper represents a creator, modifier, and deleter of entities.
type DataMapper = work.UnitDataMapper

// DataMapperFunc represents a data mapper function that performs a single
// operation, such as insert, update, or delete.
type DataMapperFunc = work.UnitDataMapperFunc
