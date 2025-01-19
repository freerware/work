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

	"github.com/avast/retry-go/v4"
	"github.com/freerware/work/v4/internal/adapters"
	"github.com/uber-go/tally/v4"
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
	Register(context.Context, ...interface{}) error

	// Cached provides the entities that have been previously registered
	// and have not been acted on via Add, Alter, or Remove.
	Cached() *UnitCache

	// Add marks the provided entities as new additions.
	Add(context.Context, ...interface{}) error

	// Alter marks the provided entities as modifications.
	Alter(context.Context, ...interface{}) error

	// Remove marks the provided entities as removals.
	Remove(context.Context, ...interface{}) error

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
	logger          UnitLogger
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
		logger:             adapters.NewNopLogger(),
		scope:              tally.NoopScope,
		actions:            make(map[UnitActionType][]UnitAction),
		retryAttempts:      3,
		retryType:          UnitRetryDelayTypeFixed,
		retryDelay:         50 * time.Millisecond,
		retryMaximumJitter: 50 * time.Millisecond,
		cacheClient:        &memoryCacheClient{},
	}
	// apply options.
	for _, opt := range options {
		opt(&o)
	}
	if !o.disableDefaultLoggingActions {
		UnitDefaultLoggingActions()(&o)
	}
	// prepare metrics scope.
	o.scope = o.scope.SubScope("unit")
	if o.db != nil {
		o.scope = o.scope.Tagged(sqlUnitTag)
	} else {
		o.scope = o.scope.Tagged(bestEffortUnitTag)
	}
	return o
}

func NewUnit(opts ...UnitOption) (Unit, error) {
	options := options(opts)
	retryOptions := []retry.Option{
		retry.Attempts(uint(options.retryAttempts)),
		retry.Delay(options.retryDelay),
		retry.DelayType(options.retryType.convert()),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(attempt uint, err error) {
			options.logger.Warn("attempted retry", "attempt", int(attempt+1), "error", err)
			options.scope.Counter(retryAttempt).Inc(1)
		}),
	}
	u := unit{
		additions:    make(map[TypeName][]interface{}),
		alterations:  make(map[TypeName][]interface{}),
		removals:     make(map[TypeName][]interface{}),
		registered:   make(map[TypeName][]interface{}),
		cached:       &UnitCache{cc: options.cacheClient, scope: options.scope},
		logger:       options.logger,
		scope:        options.scope,
		actions:      options.actions,
		db:           options.db,
		insertFuncs:  options.iFuncs(),
		updateFuncs:  options.uFuncs(),
		deleteFuncs:  options.dFuncs(),
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

func (u *unit) Register(ctx context.Context, entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeRegister)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasDeleteFunc(t) && !u.hasInsertFunc(t) && !u.hasUpdateFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), "typeName", t.String())
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.registered[t]; !ok {
			u.registered[t] = []interface{}{}
		}
		u.registered[t] = append(u.registered[t], entity)
		if cacheErr := u.cached.store(ctx, entity); cacheErr != nil {
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

func (u *unit) Add(ctx context.Context, entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeAdd)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasDeleteFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), "typeName", t.String())
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

func (u *unit) Alter(ctx context.Context, entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeAlter)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasUpdateFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), "typeName", t.String())
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.alterations[t]; !ok {
			u.alterations[t] = []interface{}{}
		}
		u.alterations[t] = append(u.alterations[t], entity)
		u.alterationCount = u.alterationCount + 1
		if err = u.cached.delete(ctx, entity); err != nil {
			u.mutex.Unlock()
			return
		}
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAlter)
	return
}

func (u *unit) Remove(ctx context.Context, entities ...interface{}) (err error) {
	u.executeActions(UnitActionTypeBeforeRemove)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if !u.hasDeleteFunc(t) {
			u.logger.Error(ErrMissingDataMapper.Error(), "typeName", t.String())
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.removals[t]; !ok {
			u.removals[t] = []interface{}{}
		}
		u.removals[t] = append(u.removals[t], entity)
		u.removalCount = u.removalCount + 1
		if err = u.cached.delete(ctx, entity); err != nil {
			u.mutex.Unlock()
			return
		}
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
