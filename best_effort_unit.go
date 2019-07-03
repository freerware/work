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
	bestEffortUnitTag map[string]string = map[string]string{
		"unit_type": "best_effort",
	}
)

type bestEffortUnit struct {
	unit

	successfulInserts map[TypeName][]interface{}
	successfulUpdates map[TypeName][]interface{}
	successfulDeletes map[TypeName][]interface{}

	successfulInsertCount int
	successfulUpdateCount int
	successfulDeleteCount int
}

// NewBestEffortUnit constructs a work unit that when faced
// with adversity, attempts rollback a single time.
func NewBestEffortUnit(parameters UnitParameters) Unit {
	u := bestEffortUnit{
		unit:              newUnit(parameters),
		successfulInserts: make(map[TypeName][]interface{}),
		successfulUpdates: make(map[TypeName][]interface{}),
		successfulDeletes: make(map[TypeName][]interface{}),
	}

	if u.hasScope() {
		u.scope = u.scope.Tagged(bestEffortUnitTag)
	}
	return &u
}

func (u *bestEffortUnit) rollbackInserts() (err error) {

	//delete successfully inserted entities.
	u.logDebug("attempting to rollback inserted entities",
		zap.Int("count", u.successfulInsertCount))
	for typeName, inserts := range u.successfulInserts {
		if err = u.deleters[typeName].Delete(inserts...); err != nil {
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return nil
}

func (u *bestEffortUnit) rollbackUpdates() (err error) {

	//reapply previously registered state for the entities.
	u.logDebug("attempting to rollback updated entities",
		zap.Int("count", u.successfulUpdateCount))
	for typeName, r := range u.registered {
		if err = u.updaters[typeName].Update(r...); err != nil {
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *bestEffortUnit) rollbackDeletes() (err error) {

	//reinsert successfully deleted entities.
	u.logDebug("attempting to rollback deleted entities",
		zap.Int("count", u.successfulDeleteCount))
	for typeName, deletes := range u.successfulDeletes {
		if err = u.inserters[typeName].Insert(deletes...); err != nil {
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}
	return
}

func (u *bestEffortUnit) rollback() (err error) {

	//setup timer.
	if u.hasScope() {
		stopWatch := u.scope.Timer("rollback").Start()
		defer func() {
			stopWatch.Stop()
			if err != nil {
				u.scope.Counter("rollback.failure").Inc(1)
			} else {
				u.scope.Counter("rollback.success").Inc(1)
			}
		}()
	}

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

	u.logDebug("attempting to insert entities", zap.Int("count", len(u.additions)))
	for typeName, additions := range u.additions {
		if err = u.inserters[typeName].Insert(additions...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
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

	u.logDebug("attempting to update entities", zap.Int("count", len(u.alterations)))
	for typeName, alterations := range u.alterations {
		if err = u.updaters[typeName].Update(alterations...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
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

	u.logDebug("attempting to remove entities", zap.Int("count", len(u.removals)))
	for typeName, removals := range u.removals {
		if err = u.deleters[typeName].Delete(removals...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
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

func (u *bestEffortUnit) Save() (err error) {

	//setup timer.
	var stopWatch tally.Stopwatch
	if u.hasScope() {
		stopWatch = u.scope.Timer("save").Start()
		defer func() {
			stopWatch.Stop()
			if err == nil {
				u.scope.Counter("save.success").Inc(1)
			}
		}()
	}

	//rollback if there is a panic.
	defer func() {
		if r := recover(); r != nil {
			err = multierr.Combine(
				fmt.Errorf("panic: unable to save work unit\n%v", r), u.rollback())
			u.logError("panic: unable to save work unit",
				zap.String("panic", fmt.Sprintf("%v", r)))
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

	totalCount := u.additionCount + u.alterationCount + u.removalCount
	u.logInfo("successfully saved unit",
		zap.Int("insertCount", u.additionCount),
		zap.Int("updateCount", u.alterationCount),
		zap.Int("deleteCount", u.removalCount),
		zap.Int("totalCount", totalCount))
	return
}
