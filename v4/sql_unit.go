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

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/avast/retry-go"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	sqlUnitTag = map[string]string{
		"unit_type": "sql",
	}
)

type sqlUnit struct {
	unit
}

func (u *sqlUnit) rollback(tx *sql.Tx) (err error) {

	//setup timer.
	stop := u.scope.Timer(rollback).Start().Stop

	//log and capture metrics.
	defer func() {
		stop()
		if err != nil {
			u.scope.Counter(rollbackFailure).Inc(1)
		} else {
			u.scope.Counter(rollbackSuccess).Inc(1)
		}
	}()
	err = tx.Rollback()
	return
}

func (u *sqlUnit) applyInserts(ctx context.Context, mCtx MapperContext) (err error) {
	for typeName, additions := range u.additions {
		var m DataMapper
		m, err = u.mapper(typeName)
		if err != nil {
			return
		}
		if err = m.Insert(ctx, mCtx, additions...); err != nil {
			u.executeActions(UnitActionTypeBeforeRollback)
			var errRb error
			if errRb = u.rollback(mCtx.Tx); errRb == nil {
				u.executeActions(UnitActionTypeAfterRollback)
			}
			err = multierr.Combine(err, errRb)
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) applyUpdates(ctx context.Context, mCtx MapperContext) (err error) {
	for typeName, alterations := range u.alterations {
		var m DataMapper
		m, err = u.mapper(typeName)
		if err != nil {
			return
		}
		if err = m.Update(ctx, mCtx, alterations...); err != nil {
			u.executeActions(UnitActionTypeBeforeRollback)
			var errRb error
			if errRb = u.rollback(mCtx.Tx); errRb == nil {
				u.executeActions(UnitActionTypeAfterRollback)
			}
			err = multierr.Combine(err, errRb)
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) applyDeletes(ctx context.Context, mCtx MapperContext) (err error) {
	for typeName, removals := range u.removals {
		var m DataMapper
		m, err = u.mapper(typeName)
		if err != nil {
			return
		}
		if err = m.Delete(ctx, mCtx, removals...); err != nil {
			u.executeActions(UnitActionTypeBeforeRollback)
			var errRb error
			if errRb = u.rollback(mCtx.Tx); errRb == nil {
				u.executeActions(UnitActionTypeAfterRollback)
			}
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) save(ctx context.Context) (err error) {
	u.retryOptions = append(u.retryOptions, retry.Context(ctx))

	//start transaction.
	tx, err := u.db.BeginTx(ctx, nil)
	mCtx := MapperContext{Tx: tx}
	if err != nil {
		// consider a failure to begin transaction as successful rollback,
		// since none of the desired changes are applied.
		u.scope.Counter(rollbackSuccess).Inc(1)
		u.logger.Error(err.Error())
		return
	}

	//rollback if there is a panic.
	defer func() {
		if r := recover(); r != nil {
			u.executeActions(UnitActionTypeBeforeRollback)
			if err = u.rollback(tx); err == nil {
				u.executeActions(UnitActionTypeAfterRollback)
			}
			msg := "panic: unable to save work unit"
			err = multierr.Combine(fmt.Errorf("%s\n%v", msg, r), err)
			u.logger.Error(msg, zap.String("panic", fmt.Sprintf("%v", r)))
			panic(r)
		}
	}()

	//insert newly added entities.
	u.executeActions(UnitActionTypeBeforeInserts)
	if err = u.applyInserts(ctx, mCtx); err != nil {
		return
	}
	u.executeActions(UnitActionTypeAfterInserts)

	//update altered entities.
	u.executeActions(UnitActionTypeBeforeUpdates)
	if err = u.applyUpdates(ctx, mCtx); err != nil {
		return
	}
	u.executeActions(UnitActionTypeAfterUpdates)

	//delete removed entities.
	u.executeActions(UnitActionTypeBeforeDeletes)
	if err = u.applyDeletes(ctx, mCtx); err != nil {
		return
	}
	u.executeActions(UnitActionTypeAfterDeletes)

	if err = tx.Commit(); err != nil {
		// consider error during transaction commit as successful rollback,
		// since the rollback is implicitly done.
		// please see https://golang.org/src/database/sql/sql.go#L1991 for reference.
		u.executeActions(UnitActionTypeAfterRollback)
		u.scope.Counter(rollbackSuccess).Inc(1)
		u.logger.Error(err.Error())
		return
	}
	return
}

// Save commits the new additions, modifications, and removals
// within the work unit to an SQL store.
func (u *sqlUnit) Save(ctx context.Context) (err error) {
	u.executeActions(UnitActionTypeBeforeSave)

	//setup timer.
	stop := u.scope.Timer(save).Start().Stop
	defer func() {
		stop()
		if err == nil {
			u.scope.Counter(saveSuccess).Inc(1)
			u.executeActions(UnitActionTypeAfterSave)
		}
	}()

	err = retry.Do(func() error { return u.save(ctx) }, u.retryOptions...)
	return
}
