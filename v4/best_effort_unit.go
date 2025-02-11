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
	"fmt"

	"github.com/avast/retry-go/v4"
	"go.uber.org/multierr"
)

var (
	bestEffortUnitTag = map[string]string{
		"unit_type": "best_effort",
	}
)

type bestEffortUnit struct {
	unit

	successfulInserts     map[TypeName][]interface{}
	successfulUpdates     map[TypeName][]interface{}
	successfulDeletes     map[TypeName][]interface{}
	successfulInsertCount int
	successfulUpdateCount int
	successfulDeleteCount int
}

func (u *bestEffortUnit) rollbackInserts(ctx context.Context, mCtx UnitMapperContext) (err error) {
	//delete successfully inserted entities.
	u.logger.Debug("attempting to rollback inserted entities", "count", u.successfulInsertCount)
	for typeName, i := range u.successfulInserts {
		if f, ok := u.deleteFunc(typeName); ok {
			if err = f(ctx, mCtx, i...); err != nil {
				u.logger.Error(err.Error(), "typeName", typeName.String())
				return
			}
		}
	}
	return nil
}

func (u *bestEffortUnit) rollbackUpdates(ctx context.Context, mCtx UnitMapperContext) (err error) {
	//reapply previously registered state for the entities.
	u.logger.Debug("attempting to rollback updated entities", "count", u.successfulUpdateCount)
	for typeName, r := range u.registered {
		if f, ok := u.updateFunc(typeName); ok {
			if err = f(ctx, mCtx, r...); err != nil {
				u.logger.Error(err.Error(), "typeName", typeName.String())
				return
			}
		}
	}
	return
}

func (u *bestEffortUnit) rollbackDeletes(ctx context.Context, mCtx UnitMapperContext) (err error) {
	//reinsert successfully deleted entities.
	u.logger.Debug("attempting to rollback deleted entities", "count", u.successfulDeleteCount)
	for typeName, d := range u.successfulDeletes {
		if f, ok := u.insertFunc(typeName); ok {
			if err = f(ctx, mCtx, d...); err != nil {
				u.logger.Error(err.Error(), "typeName", typeName.String())
				return
			}
		}
	}
	return
}

func (u *bestEffortUnit) rollback(ctx context.Context, mCtx UnitMapperContext) (err error) {
	//setup timer.
	stop := u.scope.Timer(rollback).Start().Stop

	//log and capture metrics if there is a panic.
	defer func() {
		stop()
		if r := recover(); r != nil {
			msg := "panic: unable to rollback work unit"
			u.logger.Error(msg, "panic", fmt.Sprintf("%v", r))
			u.scope.Counter(rollbackFailure).Inc(1)
			panic(r)
		}

		if err != nil {
			u.scope.Counter(rollbackFailure).Inc(1)
		} else {
			u.scope.Counter(rollbackSuccess).Inc(1)
		}
	}()

	if err = u.rollbackDeletes(ctx, mCtx); err != nil {
		return
	}

	if err = u.rollbackUpdates(ctx, mCtx); err != nil {
		return
	}

	if err = u.rollbackInserts(ctx, mCtx); err != nil {
		return
	}
	return
}

func (u *bestEffortUnit) applyInserts(ctx context.Context, mCtx UnitMapperContext) (err error) {
	for typeName, additions := range u.additions {
		if f, ok := u.insertFunc(typeName); ok {
			if err = f(ctx, mCtx, additions...); err != nil {
				u.executeActions(UnitActionTypeBeforeRollback)
				errRollback := u.rollback(ctx, mCtx)
				if errRollback == nil {
					u.executeActions(UnitActionTypeAfterRollback)
				}
				err = multierr.Combine(err, errRollback)
				u.logger.Error(err.Error(), "typeName", typeName.String())
				return
			}
			if _, ok := u.successfulInserts[typeName]; !ok {
				u.successfulInserts[typeName] = []interface{}{}
			}
			u.successfulInserts[typeName] =
				append(u.successfulInserts[typeName], additions...)
			u.successfulInsertCount = u.successfulInsertCount + len(additions)
		}
	}
	return
}

func (u *bestEffortUnit) applyUpdates(ctx context.Context, mCtx UnitMapperContext) (err error) {
	for typeName, alterations := range u.alterations {
		if f, ok := u.updateFunc(typeName); ok {
			if err = f(ctx, mCtx, alterations...); err != nil {
				u.executeActions(UnitActionTypeBeforeRollback)
				errRollback := u.rollback(ctx, mCtx)
				if errRollback == nil {
					u.executeActions(UnitActionTypeAfterRollback)
				}
				err = multierr.Combine(err, errRollback)
				u.logger.Error(err.Error(), "typeName", typeName.String())
				return
			}
			if _, ok := u.successfulUpdates[typeName]; !ok {
				u.successfulUpdates[typeName] = []interface{}{}
			}
			u.successfulUpdates[typeName] =
				append(u.successfulUpdates[typeName], alterations...)
			u.successfulUpdateCount = u.successfulUpdateCount + len(alterations)
		}
	}
	return
}

func (u *bestEffortUnit) applyDeletes(ctx context.Context, mCtx UnitMapperContext) (err error) {
	for typeName, removals := range u.removals {
		if f, ok := u.deleteFunc(typeName); ok {
			if err = f(ctx, mCtx, removals...); err != nil {
				u.executeActions(UnitActionTypeBeforeRollback)
				errRollback := u.rollback(ctx, mCtx)
				if errRollback == nil {
					u.executeActions(UnitActionTypeAfterRollback)
				}
				err = multierr.Combine(err, errRollback)
				u.logger.Error(err.Error(), "typeName", typeName.String())
				return
			}
			if _, ok := u.successfulDeletes[typeName]; !ok {
				u.successfulDeletes[typeName] = []interface{}{}
			}
			u.successfulDeletes[typeName] =
				append(u.successfulDeletes[typeName], removals...)
			u.successfulDeleteCount = u.successfulDeleteCount + len(removals)
		}
	}
	return
}

func (u *bestEffortUnit) resetSuccesses() {
	u.successfulInserts = make(map[TypeName][]interface{})
	u.successfulUpdates = make(map[TypeName][]interface{})
	u.successfulDeletes = make(map[TypeName][]interface{})
}

func (u *bestEffortUnit) resetSuccessCounts() {
	u.successfulInsertCount = 0
	u.successfulUpdateCount = 0
	u.successfulDeleteCount = 0
}

func (u *bestEffortUnit) save(ctx context.Context) (err error) {
	//insert newly added entities.
	u.executeActions(UnitActionTypeBeforeInserts)
	if err = u.applyInserts(ctx, UnitMapperContext{}); err != nil {
		return
	}
	u.executeActions(UnitActionTypeAfterInserts)

	//update altered entities.
	u.executeActions(UnitActionTypeBeforeUpdates)
	if err = u.applyUpdates(ctx, UnitMapperContext{}); err != nil {
		return
	}
	u.executeActions(UnitActionTypeAfterUpdates)

	//delete removed entities.
	u.executeActions(UnitActionTypeBeforeDeletes)
	if err = u.applyDeletes(ctx, UnitMapperContext{}); err != nil {
		return
	}
	u.executeActions(UnitActionTypeAfterDeletes)
	return
}

// Save commits the new additions, modifications, and removals
// within the work unit to a persistent store.
func (u *bestEffortUnit) Save(ctx context.Context) (err error) {
	u.executeActions(UnitActionTypeBeforeSave)

	//setup timer.
	stop := u.scope.Timer(save).Start().Stop

	//rollback if there is a panic.
	defer func() {
		stop()
		if r := recover(); r != nil {
			u.executeActions(UnitActionTypeBeforeRollback)
			if err = u.rollback(ctx, UnitMapperContext{}); err == nil {
				u.executeActions(UnitActionTypeAfterRollback)
			}
			err = multierr.Combine(
				fmt.Errorf("panic: unable to save work unit\n%v", r), err)
			u.logger.Error("panic: unable to save work unit", "panic", fmt.Sprintf("%v", r))
			panic(r)
		}
		if err == nil {
			u.scope.Counter(saveSuccess).Inc(1)
			u.scope.Counter(insert).Inc(int64(u.additionCount))
			u.scope.Counter(update).Inc(int64(u.alterationCount))
			u.scope.Counter(delete).Inc(int64(u.removalCount))
			u.executeActions(UnitActionTypeAfterSave)
		}
	}()

	onRetry :=
		retry.OnRetry(func(attempt uint, err error) {
			u.resetSuccesses()
			u.resetSuccessCounts()
			u.logger.Warn("attempted retry", "attempt", int(attempt+1), "error", err.Error())
			u.scope.Counter(retryAttempt).Inc(1)
		})
	u.retryOptions = append(u.retryOptions, retry.Context(ctx), onRetry)
	err = retry.Do(func() error { return u.save(ctx) }, u.retryOptions...)
	return
}
