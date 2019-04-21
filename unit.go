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

	"go.uber.org/zap"
)

// Unit represents an atomic set of entity changes.
type Unit interface {

	// Register tracks the provided entities as clean.
	Register(...interface{})

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
	inserters       map[TypeName]Inserter
	updaters        map[TypeName]Updater
	deleters        map[TypeName]Deleter
	additions       map[TypeName][]interface{}
	alterations     map[TypeName][]interface{}
	removals        map[TypeName][]interface{}
	registered      map[TypeName][]interface{}
	additionCount   int
	alterationCount int
	removalCount    int
	logger          *zap.Logger
}

func newUnit(parameters UnitParameters) unit {
	u := unit{
		inserters:   parameters.Inserters,
		updaters:    parameters.Updaters,
		deleters:    parameters.Deleters,
		additions:   make(map[TypeName][]interface{}),
		alterations: make(map[TypeName][]interface{}),
		removals:    make(map[TypeName][]interface{}),
		registered:  make(map[TypeName][]interface{}),
		logger:      parameters.Logger,
	}
	return u
}

func (u *unit) hasLogger() bool {
	return u.logger != nil
}

func (u *unit) logError(message string, fields ...zap.Field) {
	if u.hasLogger() {
		u.logger.Error(message, fields...)
	}
}

func (u *unit) logInfo(message string, fields ...zap.Field) {
	if u.hasLogger() {
		u.logger.Info(message, fields...)
	}
}

func (u *unit) logDebug(message string, fields ...zap.Field) {
	if u.hasLogger() {
		u.logger.Debug(message, fields...)
	}
}

// Register tracks the provided entities as clean.
func (u *unit) Register(entities ...interface{}) {
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		u.registered[tName] =
			append(u.registered[tName], entity)
	}
}

// Add marks the provided entities as new additions.
func (u *unit) Add(entities ...interface{}) error {
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if _, ok := u.inserters[tName]; !ok {
			u.logError("missing inserter", zap.String("typeName", tName.String()))
			return fmt.Errorf("missing inserter for entities with type %s", tName)
		}

		if _, ok := u.additions[tName]; !ok {
			u.additions[tName] = []interface{}{}
		}
		u.additions[tName] = append(u.additions[tName], entity)
		u.additionCount = u.additionCount + 1
	}
	return nil
}

// Alter marks the provided entities as modifications.
func (u *unit) Alter(entities ...interface{}) error {
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if _, ok := u.updaters[tName]; !ok {
			u.logError("missing updater", zap.String("typeName", tName.String()))
			return fmt.Errorf("missing updater for entities with type %s", tName)
		}

		if _, ok := u.alterations[tName]; !ok {
			u.alterations[tName] = []interface{}{}
		}
		u.alterations[tName] = append(u.alterations[tName], entity)
		u.alterationCount = u.alterationCount + 1
	}
	return nil
}

// Remove marks the provided entities as removals.
func (u *unit) Remove(entities ...interface{}) error {
	for _, entity := range entities {
		tName := TypeNameOf(entity)
		if _, ok := u.deleters[tName]; !ok {
			u.logError("missing updater", zap.String("typeName", tName.String()))
			return fmt.Errorf("missing deleter for entities with type %s", tName)
		}

		if _, ok := u.removals[tName]; !ok {
			u.removals[tName] = []interface{}{}
		}
		u.removals[tName] = append(u.removals[tName], entity)
		u.removalCount = u.removalCount + 1
	}
	return nil
}
