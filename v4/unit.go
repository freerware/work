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
	"github.com/uber-go/tally"
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
	Register(...Entity) error

	// Cached provides the entities that have been previously registered
	// and have not been acted on via Add, Alter, or Remove.
	Cached() map[TypeName][]Entity

	// Add marks the provided entities as new additions.
	Add(...Entity) error

	// Alter marks the provided entities as modifications.
	Alter(...Entity) error

	// Remove marks the provided entities as removals.
	Remove(...Entity) error

	// Save commits the new additions, modifications, and removals
	// within the work unit to a persistent store.
	Save(context.Context) error
}

type unit struct {
	additions       map[TypeName][]Entity
	alterations     map[TypeName][]Entity
	removals        map[TypeName][]Entity
	registered      map[TypeName][]Entity
	cached          map[TypeName][]Entity
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
		additions:    make(map[TypeName][]Entity),
		alterations:  make(map[TypeName][]Entity),
		removals:     make(map[TypeName][]Entity),
		registered:   make(map[TypeName][]Entity),
		cached:       make(map[TypeName][]Entity),
		logger:       options.Logger,
		scope:        options.Scope,
		actions:      options.Actions,
		db:           options.DB,
		mappers:      options.DataMappers,
		retryOptions: retryOptions,
	}
	if len(u.mappers) == 0 {
		return nil, ErrNoDataMapper
	}
	if u.db != nil {
		return &sqlUnit{unit: u}, nil
	}
	return &bestEffortUnit{
		unit:              u,
		successfulInserts: make(map[TypeName][]Entity),
		successfulUpdates: make(map[TypeName][]Entity),
		successfulDeletes: make(map[TypeName][]Entity),
	}, nil
}

func (u *unit) Register(entities ...Entity) (err error) {
	u.executeActions(UnitActionTypeBeforeRegister)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(t); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.registered[t]; !ok {
			u.registered[t] = []Entity{}
		}
		u.registered[t] = append(u.registered[t], entity)
		u.cached[t] = append(u.cached[t], entity)
		u.registerCount = u.registerCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRegister)
	return
}

func (u *unit) Cached() map[TypeName][]Entity {
	return u.cached
}

func (u *unit) Add(entities ...Entity) (err error) {
	u.executeActions(UnitActionTypeBeforeAdd)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(TypeNameOf(entity)); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.additions[t]; !ok {
			u.additions[t] = []Entity{}
		}
		u.additions[t] = append(u.additions[t], entity)
		u.additionCount = u.additionCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAdd)
	return
}

func (u *unit) Alter(entities ...Entity) (err error) {
	u.executeActions(UnitActionTypeBeforeAlter)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(TypeNameOf(entity)); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.alterations[t]; !ok {
			u.alterations[t] = []Entity{}
		}
		u.alterations[t] = append(u.alterations[t], entity)
		u.alterationCount = u.alterationCount + 1
		u.invalidate(t, entity)
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAlter)
	return
}

func (u *unit) Remove(entities ...Entity) (err error) {
	u.executeActions(UnitActionTypeBeforeRemove)
	for _, entity := range entities {
		t := TypeNameOf(entity)
		if _, err = u.mapper(TypeNameOf(entity)); err != nil {
			return
		}

		u.mutex.Lock()
		if _, ok := u.removals[t]; !ok {
			u.removals[t] = []Entity{}
		}
		u.removals[t] = append(u.removals[t], entity)
		u.removalCount = u.removalCount + 1
		u.invalidate(t, entity)
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

func (u *unit) invalidate(t TypeName, entity Entity) {
	if entities, ok := u.cached[t]; ok && len(entities) > 0 {
		cached := []Entity{}
		for _, cachedEntity := range entities {
			if cachedEntity.Identifier() != entity.Identifier() {
				cached = append(cached, cachedEntity)
			}
		}

		u.cached[t] = cached
	}
}
