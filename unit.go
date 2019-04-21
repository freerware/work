package work

import (
	"fmt"

	"go.uber.org/zap"
)

// Unit represents an atomic set of entity changes.
type Unit interface {

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
	inserters   map[TypeName]Inserter
	updaters    map[TypeName]Updater
	deleters    map[TypeName]Deleter
	additions   map[TypeName][]interface{}
	alterations map[TypeName][]interface{}
	removals    map[TypeName][]interface{}
	logger      *zap.Logger
}

func newUnit(parameters UnitParameters) unit {
	u := unit{
		inserters:   parameters.Inserters,
		updaters:    parameters.Updaters,
		deleters:    parameters.Deleters,
		additions:   make(map[TypeName][]interface{}),
		alterations: make(map[TypeName][]interface{}),
		removals:    make(map[TypeName][]interface{}),
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
	}
	return nil
}
