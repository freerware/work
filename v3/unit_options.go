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

// UnitOptions represents the configuration options
// for the work unit.
type UnitOptions struct {
	Logger                       *zap.Logger
	Scope                        tally.Scope
	Actions                      map[UnitActionType][]UnitAction
	DisableDefaultLoggingActions bool
}

// Option applies an option to the provided configuration.
type Option func(*UnitOptions)

var (
	// UnitLogger specifies the option to provide a logger for the work unit.
	UnitLogger = func(l *zap.Logger) Option {
		return func(o *UnitOptions) {
			o.Logger = l
		}
	}

	// UnitScope specifies the option to provide a metric scope for the work unit.
	UnitScope = func(s tally.Scope) Option {
		return func(o *UnitOptions) {
			o.Scope = s
		}
	}

	// setActions appends the provided actions as the provided action type.
	setActions = func(t UnitActionType, a ...UnitAction) Option {
		return func(o *UnitOptions) {
			if o.Actions == nil {
				o.Actions = make(map[UnitActionType][]UnitAction)
			}
			o.Actions[t] = append(o.Actions[t], a...)
		}
	}

	// UnitAfterRegisterActions specifies the option to provide actions to execute
	// after entities are registered with the work unit.
	UnitAfterRegisterActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterRegister, a...)
	}

	// UnitAfterAddActions specifies the option to provide actions to execute
	// after entities are added with the work unit.
	UnitAfterAddActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterAdd, a...)
	}

	// UnitAfterAlterActions specifies the option to provide actions to execute
	// after entities are altered with the work unit.
	UnitAfterAlterActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterAlter, a...)
	}

	// UnitAfterRemoveActions specifies the option to provide actions to execute
	// after entities are removed with the work unit.
	UnitAfterRemoveActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterRemove, a...)
	}

	// UnitAfterInsertsActions specifies the option to provide actions to execute
	// after new entities are inserted in the data store.
	UnitAfterInsertsActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterInserts, a...)
	}

	// UnitAfterUpdatesActions specifies the option to provide actions to execute
	// after altered entities are updated in the data store.
	UnitAfterUpdatesActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterUpdates, a...)
	}

	// UnitAfterDeletesActions specifies the option to provide actions to execute
	// after removed entities are deleted in the data store.
	UnitAfterDeletesActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterDeletes, a...)
	}

	// UnitAfterRollbackActions specifies the option to provide actions to execute
	// after a rollback is performed.
	UnitAfterRollbackActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterRollback, a...)
	}

	// UnitAfterSaveActions specifies the option to provide actions to execute
	// after a save is performed.
	UnitAfterSaveActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeAfterSave, a...)
	}

	// UnitBeforeInsertsActions specifies the option to provide actions to execute
	// before new entities are inserted in the data store.
	UnitBeforeInsertsActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeBeforeInserts, a...)
	}

	// UnitBeforeUpdatesActions specifies the option to provide actions to execute
	// before altered entities are updated in the data store.
	UnitBeforeUpdatesActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeBeforeUpdates, a...)
	}

	// UnitBeforeDeletesActions specifies the option to provide actions to execute
	// before removed entities are deleted in the data store.
	UnitBeforeDeletesActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeBeforeDeletes, a...)
	}

	// UnitBeforeRollbackActions specifies the option to provide actions to execute
	// before a rollback is performed.
	UnitBeforeRollbackActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeBeforeRollback, a...)
	}

	// UnitBeforeSaveActions specifies the option to provide actions to execute
	// before a save is performed.
	UnitBeforeSaveActions = func(a ...UnitAction) Option {
		return setActions(UnitActionTypeBeforeSave, a...)
	}

	// UnitDefaultLoggingActions specifies all of the default logging actions.
	UnitDefaultLoggingActions = func() Option {
		beforeInsertLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug(
				"attempting to insert entities",
				zap.Int("count", ctx.AdditionCount),
			)
		}
		afterInsertLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug(
				"successfully inserted entities",
				zap.Int("count", ctx.AdditionCount),
			)
		}
		beforeUpdateLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug(
				"attempting to update entities",
				zap.Int("count", ctx.AlterationCount),
			)
		}
		afterUpdateLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug(
				"successfully updated entities",
				zap.Int("count", ctx.AlterationCount),
			)
		}
		beforeDeleteLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug(
				"attempting to delete entities",
				zap.Int("count", ctx.RemovalCount),
			)
		}
		afterDeleteLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug(
				"successfully deleted entities",
				zap.Int("count", ctx.RemovalCount),
			)
		}
		beforeSaveLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to save unit")
		}
		afterSaveLogAction := func(ctx UnitActionContext) {
			totalCount :=
				ctx.AdditionCount + ctx.AlterationCount + ctx.RemovalCount
			ctx.Logger.Info("successfully saved unit",
				zap.Int("insertCount", ctx.AdditionCount),
				zap.Int("updateCount", ctx.AlterationCount),
				zap.Int("deleteCount", ctx.RemovalCount),
				zap.Int("registerCount", ctx.RegisterCount),
				zap.Int("totalUpdateCount", totalCount))
		}
		beforeRollbackLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Debug("attempting to roll back unit")
		}
		afterRollbackLogAction := func(ctx UnitActionContext) {
			ctx.Logger.Info("successfully rolled back unit")
		}
		return func(o *UnitOptions) {
			subOpts := []Option{
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
	DisableDefaultLoggingActions = func() Option {
		return func(o *UnitOptions) {
			o.DisableDefaultLoggingActions = true
		}
	}
)
