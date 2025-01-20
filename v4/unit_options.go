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
	"context"
	"database/sql"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/freerware/work/v4/internal/adapters"
	"github.com/sirupsen/logrus"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
)

// UnitOptions represents the configuration options for the work unit.
type UnitOptions struct {
	logger                       UnitLogger
	scope                        tally.Scope
	actions                      map[UnitActionType][]UnitAction
	disableDefaultLoggingActions bool
	db                           *sql.DB
	retryAttempts                int
	retryDelay                   time.Duration
	retryMaximumJitter           time.Duration
	retryType                    UnitRetryDelayType
	insertFuncs                  map[TypeName]UnitDataMapperFunc
	insertFuncsLen               int
	updateFuncs                  map[TypeName]UnitDataMapperFunc
	updateFuncsLen               int
	deleteFuncs                  map[TypeName]UnitDataMapperFunc
	deleteFuncsLen               int
	cacheClient                  UnitCacheClient
}

func (uo *UnitOptions) totalDataMapperFuncs() int {
	return uo.insertFuncsLen + uo.updateFuncsLen + uo.deleteFuncsLen
}

func (uo *UnitOptions) hasDataMapperFuncs() bool {
	return uo.totalDataMapperFuncs() != 0
}

func (uo *UnitOptions) iFuncs() (funcs *sync.Map) {
	if uo.insertFuncs == nil {
		return
	}

	funcs = &sync.Map{}
	for t, f := range uo.insertFuncs {
		funcs.Store(t, f)
	}
	return
}

func (uo *UnitOptions) uFuncs() (funcs *sync.Map) {
	if uo.updateFuncs == nil {
		return
	}

	funcs = &sync.Map{}
	for t, f := range uo.updateFuncs {
		funcs.Store(t, f)
	}
	return
}

func (uo *UnitOptions) dFuncs() (funcs *sync.Map) {
	if uo.deleteFuncs == nil {
		return
	}

	funcs = &sync.Map{}
	for t, f := range uo.deleteFuncs {
		funcs.Store(t, f)
	}
	return
}

// UnitOption applies an option to the provided configuration.
type UnitOption func(*UnitOptions)

// UnitRetryDelayType represents the type of retry delay to perform.
type UnitRetryDelayType int

func (t UnitRetryDelayType) convert() retry.DelayTypeFunc {
	types := map[UnitRetryDelayType]retry.DelayTypeFunc{
		UnitRetryDelayTypeFixed:   retry.FixedDelay,
		UnitRetryDelayTypeBackOff: retry.BackOffDelay,
		UnitRetryDelayTypeRandom:  retry.RandomDelay,
	}
	if converted, ok := types[t]; ok {
		return converted
	}
	return retry.FixedDelay
}

const (
	// Fixed represents a retry type that maintains a constaint delay between retry iterations.
	UnitRetryDelayTypeFixed = iota
	// BackOff represents a retry type that increases delay between retry iterations.
	UnitRetryDelayTypeBackOff
	// Random represents a retry type that utilizes a random delay between retry iterations.
	UnitRetryDelayTypeRandom
)

// UnitDataMapperFunc represents a data mapper function that performs a single
// operation, such as insert, update, or delete.
type UnitDataMapperFunc func(context.Context, UnitMapperContext, ...interface{}) error

var (
	// UnitDB specifies the option to provide the database for the work unit.
	UnitDB = func(db *sql.DB) UnitOption {
		return func(o *UnitOptions) {
			o.db = db
		}
	}

	// UnitDataMappers specifies the option to provide the data mappers for
	// the work unit.
	UnitDataMappers = func(dm map[TypeName]UnitDataMapper) UnitOption {
		return func(o *UnitOptions) {
			if dm == nil || len(dm) == 0 {
				return
			}
			if o.insertFuncs == nil {
				o.insertFuncs = make(map[TypeName]UnitDataMapperFunc)
			}
			if o.updateFuncs == nil {
				o.updateFuncs = make(map[TypeName]UnitDataMapperFunc)
			}
			if o.deleteFuncs == nil {
				o.deleteFuncs = make(map[TypeName]UnitDataMapperFunc)
			}
			for typeName, dataMapper := range dm {
				o.insertFuncs[typeName] = dataMapper.Insert
				o.insertFuncsLen = o.insertFuncsLen + 1
				o.updateFuncs[typeName] = dataMapper.Update
				o.updateFuncsLen = o.updateFuncsLen + 1
				o.deleteFuncs[typeName] = dataMapper.Delete
				o.deleteFuncsLen = o.deleteFuncsLen + 1
			}
		}
	}

	// UnitWithZapLogger specifies the option to provide a Zap logger for the
	// work unit.
	UnitWithZapLogger = func(l *zap.Logger) UnitOption {
		return UnitWithLogger(adapters.NewZapLogger(l))
	}

	// UnitWithStandardLogger specifies the option to provide a logger as defined
	// in the 'log' standard library package for the work unit.
	UnitWithStandardLogger = func(l *log.Logger) UnitOption {
		return UnitWithLogger(adapters.NewStandardLogger(l))
	}

	// UnitWithStructuredLogger specifies the option to provide a structured logger as defined
	// in the 'log/slog' standard library package for the work unit.
	UnitWithStructuredLogger = func(l *slog.Logger) UnitOption {
		return UnitWithLogger(adapters.NewStructuredLogger(l))
	}

	// UnitWithLogrusLogger specifies the option to provide a Logrus logger for the work unit.
	UnitWithLogrusLogger = func(l *logrus.Logger) UnitOption {
		return UnitWithLogger(adapters.NewLogrusLogger(l))
	}

	// UnitWithLogger specifies the option to provide a custom logger for the work unit.
	UnitWithLogger = func(l UnitLogger) UnitOption {
		return func(o *UnitOptions) {
			o.logger = l
		}
	}

	// UnitTallyMetricScope specifies the option to provide a tally metric
	// scope for the work unit.
	UnitTallyMetricScope = func(s tally.Scope) UnitOption {
		return func(o *UnitOptions) {
			o.scope = s
		}
	}

	// setActions appends the provided actions as the provided action type.
	setActions = func(t UnitActionType, a ...UnitAction) UnitOption {
		return func(o *UnitOptions) {
			if o.actions == nil {
				o.actions = make(map[UnitActionType][]UnitAction)
			}
			o.actions[t] = append(o.actions[t], a...)
		}
	}

	// UnitAfterRegisterActions specifies the option to provide actions to execute
	// after entities are registered with the work unit.
	UnitAfterRegisterActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterRegister, a...)
	}

	// UnitAfterAddActions specifies the option to provide actions to execute
	// after entities are added with the work unit.
	UnitAfterAddActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterAdd, a...)
	}

	// UnitAfterAlterActions specifies the option to provide actions to execute
	// after entities are altered with the work unit.
	UnitAfterAlterActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterAlter, a...)
	}

	// UnitAfterRemoveActions specifies the option to provide actions to execute
	// after entities are removed with the work unit.
	UnitAfterRemoveActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterRemove, a...)
	}

	// UnitAfterInsertsActions specifies the option to provide actions to execute
	// after new entities are inserted in the data store.
	UnitAfterInsertsActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterInserts, a...)
	}

	// UnitAfterUpdatesActions specifies the option to provide actions to execute
	// after altered entities are updated in the data store.
	UnitAfterUpdatesActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterUpdates, a...)
	}

	// UnitAfterDeletesActions specifies the option to provide actions to execute
	// after removed entities are deleted in the data store.
	UnitAfterDeletesActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterDeletes, a...)
	}

	// UnitAfterRollbackActions specifies the option to provide actions to execute
	// after a rollback is performed.
	UnitAfterRollbackActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterRollback, a...)
	}

	// UnitAfterSaveActions specifies the option to provide actions to execute
	// after a save is performed.
	UnitAfterSaveActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeAfterSave, a...)
	}

	// UnitBeforeInsertsActions specifies the option to provide actions to execute
	// before new entities are inserted in the data store.
	UnitBeforeInsertsActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeBeforeInserts, a...)
	}

	// UnitBeforeUpdatesActions specifies the option to provide actions to execute
	// before altered entities are updated in the data store.
	UnitBeforeUpdatesActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeBeforeUpdates, a...)
	}

	// UnitBeforeDeletesActions specifies the option to provide actions to execute
	// before removed entities are deleted in the data store.
	UnitBeforeDeletesActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeBeforeDeletes, a...)
	}

	// UnitBeforeRollbackActions specifies the option to provide actions to execute
	// before a rollback is performed.
	UnitBeforeRollbackActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeBeforeRollback, a...)
	}

	// UnitBeforeSaveActions specifies the option to provide actions to execute
	// before a save is performed.
	UnitBeforeSaveActions = func(a ...UnitAction) UnitOption {
		return setActions(UnitActionTypeBeforeSave, a...)
	}

	// UnitDefaultLoggingActions specifies all of the default logging actions.
	UnitDefaultLoggingActions = func() UnitOption {
		beforeInsertLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to insert entities", "count", ctx.AdditionCount)
		}
		afterInsertLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("successfully inserted entities", "count", ctx.AdditionCount)
		}
		beforeUpdateLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to update entities", "count", ctx.AlterationCount)
		}
		afterUpdateLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("successfully updated entities", "count", ctx.AlterationCount)
		}
		beforeDeleteLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to delete entities", "count", ctx.RemovalCount)
		}
		afterDeleteLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("successfully deleted entities", "count", ctx.RemovalCount)
		}
		beforeSaveLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to save unit")
		}
		afterSaveLogAction := func(ctx UnitActionContext) {
			totalCount := ctx.AdditionCount + ctx.AlterationCount + ctx.RemovalCount
			ctx.Logger.Info("successfully saved unit",
				"insertCount", ctx.AdditionCount,
				"updateCount", ctx.AlterationCount,
				"deleteCount", ctx.RemovalCount,
				"registerCount", ctx.RegisterCount,
				"totalUpdateCount", totalCount)
		}
		beforeRollbackLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to roll back unit")
		}
		afterRollbackLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Info("successfully rolled back unit")
		}
		return func(o *UnitOptions) {
			subOpts := []UnitOption{
				setActions(UnitActionTypeBeforeInserts, beforeInsertLogAction),
				setActions(UnitActionTypeAfterInserts, afterInsertLogAction),
				setActions(UnitActionTypeBeforeUpdates, beforeUpdateLogAction),
				setActions(UnitActionTypeAfterUpdates, afterUpdateLogAction),
				setActions(UnitActionTypeBeforeDeletes, beforeDeleteLogAction),
				setActions(UnitActionTypeAfterDeletes, afterDeleteLogAction),
				setActions(UnitActionTypeBeforeSave, beforeSaveLogAction),
				setActions(UnitActionTypeAfterSave, afterSaveLogAction),
				setActions(UnitActionTypeBeforeRollback, beforeRollbackLogAction),
				setActions(UnitActionTypeAfterRollback, afterRollbackLogAction),
			}
			for _, opt := range subOpts {
				opt(o)
			}
		}
	}

	// DisableDefaultLoggingActions disables the default logging actions.
	DisableDefaultLoggingActions = func() UnitOption {
		return func(o *UnitOptions) {
			o.disableDefaultLoggingActions = true
		}
	}

	// UnitRetryAttempts defines the number of retry attempts to perform.
	UnitRetryAttempts = func(attempts int) UnitOption {
		if attempts < 0 {
			attempts = 0
		}
		return func(o *UnitOptions) {
			o.retryAttempts = attempts
		}
	}

	// UnitRetryDelay defines the delay to utilize during retries.
	UnitRetryDelay = func(delay time.Duration) UnitOption {
		return func(o *UnitOptions) {
			o.retryDelay = delay
		}
	}

	// UnitRetryMaximumJitter defines the maximum jitter to utilize during
	// retries that utilize random delay times.
	UnitRetryMaximumJitter = func(jitter time.Duration) UnitOption {
		return func(o *UnitOptions) {
			o.retryMaximumJitter = jitter
		}
	}

	// UnitRetryType defines the type of retry to perform.
	UnitRetryType = func(retryType UnitRetryDelayType) UnitOption {
		return func(o *UnitOptions) {
			o.retryType = retryType
		}
	}

	// UnitInsertFunc defines the function to be used for inserting new
	// entities in the underlying data store.
	UnitInsertFunc = func(t TypeName, insertFunc UnitDataMapperFunc) UnitOption {
		return func(o *UnitOptions) {
			if o.insertFuncs == nil {
				o.insertFuncs = make(map[TypeName]UnitDataMapperFunc)
			}
			o.insertFuncs[t] = insertFunc
			o.insertFuncsLen = o.insertFuncsLen + 1
		}
	}

	// UnitUpdateFunc defines the function to be used for updating existing
	// entities in the underlying data store.
	UnitUpdateFunc = func(t TypeName, updateFunc UnitDataMapperFunc) UnitOption {
		return func(o *UnitOptions) {
			if o.updateFuncs == nil {
				o.updateFuncs = make(map[TypeName]UnitDataMapperFunc)
			}
			o.updateFuncs[t] = updateFunc
			o.updateFuncsLen = o.updateFuncsLen + 1
		}
	}

	// UnitDeleteFunc defines the function to be used for deleting existing
	// entities in the underlying data store.
	UnitDeleteFunc = func(t TypeName, deleteFunc UnitDataMapperFunc) UnitOption {
		return func(o *UnitOptions) {
			if o.deleteFuncs == nil {
				o.deleteFuncs = make(map[TypeName]UnitDataMapperFunc)
			}
			o.deleteFuncs[t] = deleteFunc
			o.deleteFuncsLen = o.deleteFuncsLen + 1
		}
	}

	// UnitWithCacheClient defines the cache client to be used.
	UnitWithCacheClient = func(cc UnitCacheClient) UnitOption {
		return func(o *UnitOptions) {
			o.cacheClient = cc
		}
	}
)
