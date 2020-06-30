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
	"fmt"

	"github.com/uber-go/tally"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	bestEffortUnitTag = map[string]string{
		"unit_type": "best_effort",
	}
)

type bestEffortUnit struct {
	unit

	mappers               map[TypeName]DataMapper
	successfulInserts     map[TypeName][]interface{}
	successfulUpdates     map[TypeName][]interface{}
	successfulDeletes     map[TypeName][]interface{}
	successfulInsertCount int
	successfulUpdateCount int
	successfulDeleteCount int
}

// NewBestEffortUnit constructs a work unit that when faced
// with adversity, attempts rollback a single time.
func NewBestEffortUnit(
	mappers map[TypeName]DataMapper, options ...Option) (Unit, error) {
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
	o.Scope = o.Scope.Tagged(bestEffortUnitTag)

	u := bestEffortUnit{
		unit:              newUnit(o),
		mappers:           mappers,
		successfulInserts: make(map[TypeName][]interface{}),
		successfulUpdates: make(map[TypeName][]interface{}),
		successfulDeletes: make(map[TypeName][]interface{}),
	}
	return &u, nil
}

func (u *bestEffortUnit) rollbackInserts() (err error) {

	//delete successfully inserted entities.
	u.logger.Debug("attempting to rollback inserted entities",
		zap.Int("count", u.successfulInsertCount))
	for typeName, inserts := range u.successfulInserts {
		if err = u.mappers[typeName].Delete(inserts...); err != nil {
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return nil
}

func (u *bestEffortUnit) rollbackUpdates() (err error) {

	//reapply previously registered state for the entities.
	u.logger.Debug("attempting to rollback updated entities",
		zap.Int("count", u.successfulUpdateCount))
	for typeName, r := range u.registered {
		if err = u.mappers[typeName].Update(r...); err != nil {
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *bestEffortUnit) rollbackDeletes() (err error) {

	//reinsert successfully deleted entities.
	u.logger.Debug("attempting to rollback deleted entities",
		zap.Int("count", u.successfulDeleteCount))
	for typeName, deletes := range u.successfulDeletes {
		if err = u.mappers[typeName].Insert(deletes...); err != nil {
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *bestEffortUnit) rollback() (err error) {

	//setup timer.
	stop := u.scope.Timer(rollback).Start().Stop

	//log and capture metrics if there is a panic.
	defer func() {
		stop()
		if r := recover(); r != nil {
			msg := "panic: unable to rollback work unit"
			u.logger.Error(msg, zap.String("panic", fmt.Sprintf("%v", r)))
			u.scope.Counter(rollbackFailure).Inc(1)
			panic(r)
		}

		if err != nil {
			u.scope.Counter(rollbackFailure).Inc(1)
		} else {
			u.scope.Counter(rollbackSuccess).Inc(1)
		}
	}()

	if err = u.rollbackDeletes(); err != nil {
		return
	}

	if err = u.rollbackUpdates(); err != nil {
		return
	}

	if err = u.rollbackInserts(); err != nil {
		return
	}
	return
}

func (u *bestEffortUnit) applyInserts() (err error) {
	u.logger.Debug("attempting to insert entities", zap.Int("count", len(u.additions)))
	for typeName, additions := range u.additions {
		if err = u.mappers[typeName].Insert(additions...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
		if _, ok := u.successfulInserts[typeName]; !ok {
			u.successfulInserts[typeName] = []interface{}{}
		}
		u.successfulInserts[typeName] =
			append(u.successfulInserts[typeName], additions...)
		u.successfulInsertCount = u.successfulInsertCount + len(additions)
	}
	return
}

func (u *bestEffortUnit) applyUpdates() (err error) {
	u.logger.Debug("attempting to update entities", zap.Int("count", len(u.alterations)))
	for typeName, alterations := range u.alterations {
		if err = u.mappers[typeName].Update(alterations...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
		if _, ok := u.successfulUpdates[typeName]; !ok {
			u.successfulUpdates[typeName] = []interface{}{}
		}
		u.successfulUpdates[typeName] =
			append(u.successfulUpdates[typeName], alterations...)
		u.successfulUpdateCount = u.successfulUpdateCount + len(alterations)
	}
	return
}

func (u *bestEffortUnit) applyDeletes() (err error) {
	u.logger.Debug("attempting to remove entities", zap.Int("count", len(u.removals)))
	for typeName, removals := range u.removals {
		if err = u.mappers[typeName].Delete(removals...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logger.Error(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
		if _, ok := u.successfulDeletes[typeName]; !ok {
			u.successfulDeletes[typeName] = []interface{}{}
		}
		u.successfulDeletes[typeName] =
			append(u.successfulDeletes[typeName], removals...)
		u.successfulDeleteCount = u.successfulDeleteCount + len(removals)
	}
	return
}

// Register tracks the provided entities as clean.
func (u *bestEffortUnit) Register(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.register(c, entities...)
}

// Add marks the provided entities as new additions.
func (u *bestEffortUnit) Add(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.add(c, entities...)
}

// Alter marks the provided entities as modifications.
func (u *bestEffortUnit) Alter(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.alter(c, entities...)
}

// Remove marks the provided entities as removals.
func (u *bestEffortUnit) Remove(entities ...interface{}) error {
	c := func(t TypeName) bool {
		_, ok := u.mappers[t]
		return ok
	}
	return u.remove(c, entities...)
}

// Save commits the new additions, modifications, and removals
// within the work unit to a persistent store.
func (u *bestEffortUnit) Save() (err error) {

	//setup timer.
	stop := u.scope.Timer(save).Start().Stop

	//rollback if there is a panic.
	defer func() {
		stop()
		if r := recover(); r != nil {
			err = multierr.Combine(
				fmt.Errorf("panic: unable to save work unit\n%v", r), u.rollback())
			u.logger.Error("panic: unable to save work unit",
				zap.String("panic", fmt.Sprintf("%v", r)))
			panic(r)
		}
		if err == nil {
			u.scope.Counter(saveSuccess).Inc(1)
		}
	}()

	//insert newly added entities.
	if err = u.applyInserts(); err != nil {
		return
	}

	//update altered entities.
	if err = u.applyUpdates(); err != nil {
		return
	}

	//delete removed entities.
	if err = u.applyDeletes(); err != nil {
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
