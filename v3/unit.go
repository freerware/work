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
	"errors"
	"sync"

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
	Save() error
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
}

func newUnit(options UnitOptions) unit {
	if !options.DisableDefaultLoggingActions {
		UnitDefaultLoggingActions()(&options)
	}

	u := unit{
		additions:   make(map[TypeName][]interface{}),
		alterations: make(map[TypeName][]interface{}),
		removals:    make(map[TypeName][]interface{}),
		registered:  make(map[TypeName][]interface{}),
		logger:      options.Logger,
		scope:       options.Scope.SubScope("unit"),
		actions:     options.Actions,
	}
	return u
}

func (u *unit) register(checker func(t TypeName) bool, entities ...interface{}) error {
	u.executeActions(UnitActionTypeBeforeRegister)
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if ok := checker(tName); !ok {
			u.logger.Error(
				ErrMissingDataMapper.Error(), zap.String("typeName", tName.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.registered[tName]; !ok {
			u.registered[tName] = []interface{}{}
		}
		u.registered[tName] = append(u.registered[tName], entity)
		u.registerCount = u.registerCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRegister)
	return nil
}

func (u *unit) add(checker func(t TypeName) bool, entities ...interface{}) error {
	u.executeActions(UnitActionTypeBeforeAdd)
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if ok := checker(tName); !ok {
			u.logger.Error(
				ErrMissingDataMapper.Error(), zap.String("typeName", tName.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.additions[tName]; !ok {
			u.additions[tName] = []interface{}{}
		}
		u.additions[tName] = append(u.additions[tName], entity)
		u.additionCount = u.additionCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAdd)
	return nil
}

func (u *unit) alter(checker func(t TypeName) bool, entities ...interface{}) error {
	u.executeActions(UnitActionTypeBeforeAlter)
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if ok := checker(tName); !ok {
			u.logger.Error(
				ErrMissingDataMapper.Error(), zap.String("typeName", tName.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.alterations[tName]; !ok {
			u.alterations[tName] = []interface{}{}
		}
		u.alterations[tName] = append(u.alterations[tName], entity)
		u.alterationCount = u.alterationCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterAlter)
	return nil
}

func (u *unit) remove(checker func(t TypeName) bool, entities ...interface{}) error {
	u.executeActions(UnitActionTypeBeforeRemove)
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if ok := checker(tName); !ok {
			u.logger.Error(
				ErrMissingDataMapper.Error(), zap.String("typeName", tName.String()))
			return ErrMissingDataMapper
		}

		u.mutex.Lock()
		if _, ok := u.removals[tName]; !ok {
			u.removals[tName] = []interface{}{}
		}
		u.removals[tName] = append(u.removals[tName], entity)
		u.removalCount = u.removalCount + 1
		u.mutex.Unlock()
	}
	u.executeActions(UnitActionTypeAfterRemove)
	return nil
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
