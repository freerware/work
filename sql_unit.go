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
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	sqlUnitTag map[string]string = map[string]string{
		"unit_type": "sql",
	}
)

type sqlUnit struct {
	unit

	connectionPool *sql.DB
}

// NewSQLUnit constructs a work unit for SQL stores.
func NewSQLUnit(parameters SQLUnitParameters) (Unit, error) {
	if parameters.ConnectionPool == nil {
		return nil, errors.New("must provide connection pool")
	}

	u := sqlUnit{
		unit:           newUnit(parameters.UnitParameters),
		connectionPool: parameters.ConnectionPool,
	}

	if u.hasScope() {
		u.scope = u.scope.Tagged(sqlUnitTag)
	}

	return &u, nil
}

func (u *sqlUnit) rollback(tx *sql.Tx) (err error) {

	//setup timer.
	stop := u.startTimer(rollback)

	//log and capture metrics.
	defer func() {
		stop()
		if err != nil {
			u.incrementCounter(rollbackFailure, 1)
		} else {
			u.incrementCounter(rollbackSuccess, 1)
		}
	}()

	err = tx.Rollback()
	return
}

func (u *sqlUnit) applyInserts(tx *sql.Tx) (err error) {
	u.logDebug("attempting to insert entities", zap.Int("count", u.additionCount))
	for typeName, additions := range u.additions {
		if err = u.inserters[typeName].Insert(additions...); err != nil {
			err = multierr.Combine(err, u.rollback(tx))
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) applyUpdates(tx *sql.Tx) (err error) {
	u.logDebug("attempting to update entities", zap.Int("count", u.alterationCount))
	for typeName, alterations := range u.alterations {
		if err = u.updaters[typeName].Update(alterations...); err != nil {
			err = multierr.Combine(err, u.rollback(tx))
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) applyDeletes(tx *sql.Tx) (err error) {
	u.logDebug("attempting to remove entities", zap.Int("count", u.removalCount))
	for typeName, removals := range u.removals {
		if err = u.deleters[typeName].Delete(removals...); err != nil {
			err = multierr.Combine(err, u.rollback(tx))
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

// Save commits the new additions, modifications, and removals
// within the work unit to an SQL store.
func (u *sqlUnit) Save() (err error) {

	//setup timer.
	stop := u.startTimer(save)
	defer func() {
		stop()
		if err == nil {
			u.incrementCounter(saveSuccess, 1)
		}
	}()

	//start transaction.
	tx, err := u.connectionPool.Begin()
	if err != nil {
		// consider a failure to begin transaction as successful rollback,
		// since none of the desired changes are applied.
		u.incrementCounter(rollbackSuccess, 1)
		u.logError(err.Error())
		return
	}

	//rollback if there is a panic.
	defer func() {
		if r := recover(); r != nil {
			msg := "panic: unable to save work unit"
			err = multierr.Combine(fmt.Errorf("%s\n%v", msg, r), u.rollback(tx))
			u.logError(msg, zap.String("panic", fmt.Sprintf("%v", r)))
			panic(r)
		}
	}()

	//insert newly added entities.
	if err = u.applyInserts(tx); err != nil {
		return
	}

	//update altered entities.
	if err = u.applyUpdates(tx); err != nil {
		return
	}

	//delete removed entities.
	if err = u.applyDeletes(tx); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		// consider error during transaction commit as successful rollback,
		// since the rollback is implicitly done.
		// please see https://golang.org/src/database/sql/sql.go#L1991 for reference.
		u.incrementCounter(rollbackSuccess, 1)
		u.logError(err.Error())
		return
	}

	totalCount := u.additionCount + u.alterationCount + u.removalCount
	u.logInfo("successfully saved unit",
		zap.Int("insertCount", u.additionCount),
		zap.Int("updateCount", u.alterationCount),
		zap.Int("deleteCount", u.removalCount),
		zap.Int("totalCount", totalCount))
	return
}
