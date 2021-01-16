/* Copyright 2021 Freerware
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
	"fmt"

	"github.com/uber-go/tally"
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

	mappers map[TypeName]SQLDataMapper
	db      *sql.DB
}

// NewSQLUnit constructs a work unit for SQL data stores.
func NewSQLUnit(
	mappers map[TypeName]SQLDataMapper,
	db *sql.DB,
	options ...Option,
) (Unit, error) {
	// validate.
	if len(mappers) < 1 {
		return nil, ErrNoDataMapper
	}

	// set defaults.
	o := UnitOptions{
		Logger: zap.NewNop(),
		Scope:  tally.NoopScope,
	}

	// apply options.
	for _, opt := range options {
		opt(&o)
	}
	o.Scope = o.Scope.Tagged(sqlUnitTag)

	u := sqlUnit{
		unit:    newUnit(o),
		mappers: mappers,
		db:      db,
	}
	return &u, nil
}

// Register tracks the provided entities as clean.
func (u *sqlUnit) Register(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.register(c, entities...)
}

// Add marks the provided entities as new additions.
func (u *sqlUnit) Add(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.add(c, entities...)
}

// Alter marks the provided entities as modifications.
func (u *sqlUnit) Alter(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.alter(c, entities...)
}

// Remove marks the provided entities as removals.
func (u *sqlUnit) Remove(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.remove(c, entities...)
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

func (u *sqlUnit) applyInserts(tx *sql.Tx) (err error) {
	u.logger.Debug("attempting to insert entities", zap.Int("count", u.additionCount))
	for typeName, additions := range u.additions {
		if err = u.mappers[typeName].Insert(tx, additions...); err != nil {
			err = multierr.Combine(err, u.rollback(tx))
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) applyUpdates(tx *sql.Tx) (err error) {
	u.logger.Debug("attempting to update entities", zap.Int("count", u.alterationCount))
	for typeName, alterations := range u.alterations {
		if err = u.mappers[typeName].Update(tx, alterations...); err != nil {
			err = multierr.Combine(err, u.rollback(tx))
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *sqlUnit) applyDeletes(tx *sql.Tx) (err error) {
	u.logger.Debug("attempting to remove entities", zap.Int("count", u.removalCount))
	for typeName, removals := range u.removals {
		if err = u.mappers[typeName].Delete(tx, removals...); err != nil {
			err = multierr.Combine(err, u.rollback(tx))
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

// Save commits the new additions, modifications, and removals
// within the work unit to an SQL store.
func (u *sqlUnit) Save() (err error) {

	//setup timer.
	stop := u.scope.Timer(save).Start().Stop
	defer func() {
		stop()
		if err == nil {
			u.scope.Counter(saveSuccess).Inc(1)
		}
	}()

	//start transaction.
	tx, err := u.db.Begin()
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
			msg := "panic: unable to save work unit"
			err = multierr.Combine(fmt.Errorf("%s\n%v", msg, r), u.rollback(tx))
			u.logger.Error(msg, zap.String("panic", fmt.Sprintf("%v", r)))
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
		u.scope.Counter(rollbackSuccess).Inc(1)
		u.logger.Error(err.Error())
		return
	}

	totalCount :=
		u.additionCount + u.alterationCount + u.removalCount + u.registerCount
	u.logger.Info("successfully saved unit",
		zap.Int("insertCount", u.additionCount),
		zap.Int("updateCount", u.alterationCount),
		zap.Int("deleteCount", u.removalCount),
		zap.Int("registerCount", u.registerCount),
		zap.Int("totalCount", totalCount))
	return
}
