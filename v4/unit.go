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

package work

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
)

// Metric scope name definitions.
const (
	rollbackSuccess = "rollback.success"
	rollbackFailure = "rollback.failure"
	saveSuccess     = "save.success"
	save            = "save"
	rollback        = "rollback"
	retryAttempt    = "retry.attempt"
	insert          = "insert"
	update          = "update"
	delete          = "delete"
	cacheInsert     = "cache.insert"
	cacheDelete     = "cache.delete"
)

var (

	// ErrMissingDataMapper represents the error that is returned
	// when attempting to add, alter, remove, or register an entity
	// that doesn't have a corresponding data mapper.
	ErrMissingDataMapper = errors.New("missing data mapper or data mapper function for entity")

	// ErrNoDataMapper represents the error that occurs when attempting
	// to create a work unit without any data mappers.
	ErrNoDataMapper = errors.New("must have at least one data mapper or data mapper function")
)

// Unit represents an atomic set of entity changes.
type Unit interface {

	// Register tracks the provided entities as clean.
	Register(...interface{}) error

	// Cached provides the entities that have been previously registered
	// and have not been acted on via Add, Alter, or Remove.
	Cached() *UnitCache

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
	cached          *UnitCache
	additionCount   int
	alterationCount int
	removalCount    int
	registerCount   int
	logger          *zap.Logger
	scope           tally.Scope
	actions         map[UnitActionType][]UnitAction
	mutex           sync.RWMutex
	db              *sql.DB
	retryOptions    []retry.Option
	insertFuncs     *sync.Map
	updateFuncs     *sync.Map
	deleteFuncs     *sync.Map
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
	// prepare metrics scope.
	o.Scope = o.Scope.SubScope("unit")
	if o.DB != nil {
		o.Scope = o.Scope.Tagged(sqlUnitTag)
	} else {
		o.Scope = o.Scope.Tagged(bestEffortUnitTag)
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
			options.Scope.Counter(retryAttempt).Inc(1)
		}),
	}
	u := unit{
		additions:    make(map[TypeName][]interface{}),
		alterations:  make(map[TypeName][]interface{}),
		removals:     make(map[TypeName][]interface{}),
		registered:   make(map[TypeName][]interface{}),
		cached:       &UnitCache{scope: options.Scope},
		logger:       options.Logger,
		scope:        options.Scope,
		actions:      options.Actions,
		db:           options.DB,
		insertFuncs:  options.insertFuncs(),
		updateFuncs:  options.updateFuncs(),
		deleteFuncs:  options.deleteFuncs(),
		retryOptions: retryOptions,
	}
	if !options.hasDataMapperFuncs() {
		return nil, ErrNoDataMapper
	}
	if u.db != nil {
		return &sqlUnit{unit: u}, nil
	}
	return &bestEffortUnit{
		unit:              u,
		successfulInserts: make(map[TypeName][]interface{}),
		successfulUpdates: make(map[TypeName][]interface{}),
		successfulDeletes: make(map[TypeName][]interface{}),
	}, nil
}

func id(entity interface{}) (interface{}, bool) {
	switch i := entity.(type) {
	case identifierer:
		return i.Identifier(), true
	case ider:
		return i.ID(), true
	default:
		return nil, false
	}
}

func (u *unit) Register(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeRegister)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasDeleteFunc(t) && !u.hasInsertFunc(t) && !u.hasUpdateFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), zap.String("typeName", t.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.registered[t]; !ok {
			u.registered[t] = []interface{}{}
		}
		u.registered[t] = append(u.registered[t], entity)
		if cacheErr := u.cached.store(entity); cacheErr != nil {
			u.logger.Warn(cacheErr.Error())
		}
		u.registerCount = u.registerCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRegister)
	return
}

func (u *unit) Cached() *UnitCache {
	return u.cached
}

func (u *unit) Add(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeAdd)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasDeleteFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), zap.String("typeName", t.String()))
			return ErrMissingDataMapper
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
		if !u.hasUpdateFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), zap.String("typeName", t.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.alterations[t]; !ok {
			u.alterations[t] = []interface{}{}
		}
		u.alterations[t] = append(u.alterations[t], entity)
		u.alterationCount = u.alterationCount + 1
		u.cached.delete(entity)
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAlter)
	return
}

func (u *unit) Remove(entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeRemove)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasDeleteFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), zap.String("typeName", t.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.removals[t]; !ok {
			u.removals[t] = []interface{}{}
		}
		u.removals[t] = append(u.removals[t], entity)
		u.removalCount = u.removalCount + 1
		u.cached.delete(entity)
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRemove)
	return
}

func (u *unit) insertFunc(t TypeName) (f UnitDataMapperFunc, ok bool) {
	if val, exists := u.insertFuncs.Load(t); exists {
		if f, ok = val.(UnitDataMapperFunc); ok {
			return
		}
	}
	return
}

func (u *unit) hasInsertFunc(t TypeName) (ok bool) {
	_, ok = u.insertFunc(t)
	return
}

func (u *unit) updateFunc(t TypeName) (f UnitDataMapperFunc, ok bool) {
	if val, exists := u.updateFuncs.Load(t); exists {
		if f, ok = val.(UnitDataMapperFunc); ok {
			return
		}
	}
	return
}

func (u *unit) hasUpdateFunc(t TypeName) (ok bool) {
	_, ok = u.updateFunc(t)
	return
}

func (u *unit) deleteFunc(t TypeName) (f UnitDataMapperFunc, ok bool) {
	if val, exists := u.deleteFuncs.Load(t); exists {
		if f, ok = val.(UnitDataMapperFunc); ok {
			return
		}
	}
	return
}

func (u *unit) hasDeleteFunc(t TypeName) (ok bool) {
	_, ok = u.deleteFunc(t)
	return
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
