package work

import (
	"fmt"

	"go.uber.org/multierr"
	"go.uber.org/zap"
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
	return &u
}

func (u *bestEffortUnit) rollbackInserts() error {

	//delete successfully inserted entities.
	u.logDebug("attempting to rollback inserted entities",
		zap.Int("count", u.successfulInsertCount))
	for typeName, inserts := range u.successfulInserts {
		if err := u.deleters[typeName].Delete(inserts...); err != nil {
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return err
		}
	}
	return nil
}

func (u *bestEffortUnit) rollbackUpdates() error {

	//reapply previously registered state for the entities.
	u.logDebug("attempting to rollback updated entities",
		zap.Int("count", u.successfulUpdateCount))
	for typeName, r := range u.registered {
		if err := u.updaters[typeName].Update(r...); err != nil {
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return err
		}
	}
	return nil
}

func (u *bestEffortUnit) rollbackDeletes() error {

	//reinsert successfully deleted entities.
	u.logDebug("attempting to rollback deleted entities",
		zap.Int("count", u.successfulDeleteCount))
	for typeName, deletes := range u.successfulDeletes {
		if err := u.inserters[typeName].Insert(deletes...); err != nil {
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return err
		}
	}
	return nil
}

func (u *bestEffortUnit) rollback() error {
	if err := u.rollbackDeletes(); err != nil {
		return err
	}

	if err := u.rollbackUpdates(); err != nil {
		return err
	}

	return u.rollbackInserts()
}

func (u *bestEffortUnit) applyInserts() error {

	u.logDebug("attempting to insert entities", zap.Int("count", len(u.additions)))
	for typeName, additions := range u.additions {
		if err := u.inserters[typeName].Insert(additions...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return err
		}
		if _, ok := u.successfulInserts[typeName]; !ok {
			u.successfulInserts[typeName] = []interface{}{}
		}
		u.successfulInserts[typeName] =
			append(u.successfulInserts[typeName], additions...)
		u.successfulInsertCount = u.successfulInsertCount + len(additions)
	}
	return nil
}

func (u *bestEffortUnit) applyUpdates() error {

	u.logDebug("attempting to update entities", zap.Int("count", len(u.alterations)))
	for typeName, alterations := range u.alterations {
		if err := u.updaters[typeName].Update(alterations...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return err
		}
		if _, ok := u.successfulUpdates[typeName]; !ok {
			u.successfulUpdates[typeName] = []interface{}{}
		}
		u.successfulUpdates[typeName] =
			append(u.successfulUpdates[typeName], alterations...)
		u.successfulUpdateCount = u.successfulUpdateCount + len(alterations)
	}
	return nil
}

func (u *bestEffortUnit) applyDeletes() error {

	u.logDebug("attempting to remove entities", zap.Int("count", len(u.removals)))
	for typeName, removals := range u.removals {
		if err := u.deleters[typeName].Delete(removals...); err != nil {
			err = multierr.Combine(err, u.rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return err
		}
		if _, ok := u.successfulDeletes[typeName]; !ok {
			u.successfulDeletes[typeName] = []interface{}{}
		}
		u.successfulDeletes[typeName] =
			append(u.successfulDeletes[typeName], removals...)
		u.successfulDeleteCount = u.successfulDeleteCount + len(removals)
	}
	return nil
}

func (u *bestEffortUnit) Save() (err error) {

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
