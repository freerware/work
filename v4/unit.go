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
	"errors"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

const (
	rollbackSuccess = "rollback.success"
	rollbackFailure = "rollback.failure"
	saveSuccess     = "save.success"
	save            = "save"
	rollback        = "rollback"
)

var (

	// ErrMissingDataMapper represents the error that is returned
	// when attempting to add, alter, remove, or register an entity
	// that doesn't have a corresponding data mapper.
	ErrMissingDataMapper = errors.New("missing data mapper for entity")

	// ErrNoDataMapper represents the error that occurs when attempting
	// to create a work unit without any data mappers.
	ErrNoDataMapper = errors.New("must have at least one data mapper")
)

// Unit represents an atomic set of entity changes.
type Unit interface {

	// Register tracks the provided entities as clean.
	Register(...interface{}) error

	// Add marks the provided entities as new additions.
	Add(...interface{}) error

	// Alter marks the provided entities as modifications.
	Alter(...interface{}) error

	// Remove marks the provided entities as removals.
	Remove(...interface{}) error

	// Save commits the new additions, modifications, and removals
	// within the work unit to a persistent store.
	Save(context.Context) error
}

type unit struct {
	additions       map[TypeName][]interface{}
	alterations     map[TypeName][]interface{}
	removals        map[TypeName][]interface{}
	registered      map[TypeName][]interface{}
	additionCount   int
	alterationCount int
	removalCount    int
	registerCount   int
	logger          *zap.Logger
	scope           tally.Scope
	actions         map[UnitActionType][]UnitAction
	mutex           sync.RWMutex
	db              *sql.DB
	mappers         map[TypeName]DataMapper
	retryOptions    []retry.Option
}

func options(options []UnitOption) UnitOptions {
	// set defaults.
	o := UnitOptions{
		Logger:             zap.NewNop(),
		Scope:              tally.NoopScope,
		Actions:            make(map[UnitActionType][]UnitAction),
		RetryAttempts:      3,
		RetryType:          UnitRetryDelayTypeFixed,
		RetryDelay:         50 * time.Millisecond,
		RetryMaximumJitter: 50 * time.Millisecond,
	}
	// apply options.
	for _, opt := range options {
		opt(&o)
	}
	if !o.DisableDefaultLoggingActions {
		UnitDefaultLoggingActions()(&o)
	}
	return o
}

func NewUnit(opts ...UnitOption) (Unit, error) {
	options := options(opts)
	retryOptions := []retry.Option{
		retry.Attempts(uint(options.RetryAttempts)),
		retry.Delay(options.RetryDelay),
		retry.DelayType(options.RetryType.convert()),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(attempt uint, err error) {
			options.Logger.Warn(
				"attempted retry",
				zap.Int("attempt", int(attempt+1)),
				zap.Error(err),
			)
		}),
	}
	u := unit{
		additions:    make(map[TypeName][]interface{}),
		alterations:  make(map[TypeName][]interface{}),
		removals:     make(map[TypeName][]interface{}),
		registered:   make(map[TypeName][]interface{}),
		logger:       options.Logger,
		scope:        options.Scope.SubScope("unit"),
		actions:      options.Actions,
		db:           options.DB,
		mappers:      options.DataMappers,
		retryOptions: retryOptions,
	}
	if len(u.mappers) == 0 {
		return nil, ErrNoDataMapper
	}
	if u.db != nil {
		u.scope = u.scope.Tagged(sqlUnitTag)
		return &sqlUnit{unit: u}, nil
	}
	u.scope = u.scope.Tagged(bestEffortUnitTag)
	return &bestEffortUnit{
		unit:              u,
		successfulInserts: make(map[TypeName][]interface{}),
		successfulUpdates: make(map[TypeName][]interface{}),
		successfulDeletes: make(map[TypeName][]interface{}),
	}, nil
}

func (u *unit) Register(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeRegister)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(t); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.registered[t]; !ok {
			u.registered[t] = []interface{}{}
		}
		u.registered[t] = append(u.registered[t], entity)
		u.registerCount = u.registerCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRegister)
	return
}

func (u *unit) Add(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeAdd)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(TypeNameOf(entity)); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.additions[t]; !ok {
			u.additions[t] = []interface{}{}
		}
		u.additions[t] = append(u.additions[t], entity)
		u.additionCount = u.additionCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAdd)
	return
}

func (u *unit) Alter(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeAlter)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(TypeNameOf(entity)); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.alterations[t]; !ok {
			u.alterations[t] = []interface{}{}
		}
		u.alterations[t] = append(u.alterations[t], entity)
		u.alterationCount = u.alterationCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAlter)
	return
}

func (u *unit) Remove(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeRemove)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(TypeNameOf(entity)); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.removals[t]; !ok {
			u.removals[t] = []interface{}{}
		}
		u.removals[t] = append(u.removals[t], entity)
		u.removalCount = u.removalCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRemove)
	return
}

func (u *unit) mapper(t TypeName) (DataMapper, error) {
	u.mutex.RLock()
	defer func() {
		u.mutex.RUnlock()
	}()

	m, ok := u.mappers[t]
	if !ok {
		u.logger.Error(ErrMissingDataMapper.Error(), zap.String("typeName", t.String()))
		return nil, ErrMissingDataMapper
	}
	return m, nil
}

func (u *unit) executeActions(actionType UnitActionType) {
	for _, action := range u.actions[actionType] {
		action(UnitActionContext{
			Logger:          u.logger,
			Scope:           u.scope,
			AdditionCount:   u.additionCount,
			AlterationCount: u.alterationCount,
			RemovalCount:    u.removalCount,
			RegisterCount:   u.registerCount,
		})
	}
}
