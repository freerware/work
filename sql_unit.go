package work

import (
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type sqlUnit struct {
	unit

	connectionPool *sql.DB
}

// NewSQLUnit constructs a work unit for SQL stores.
func NewSQLUnit(parameters SQLUnitParameters) (Unit, error) {
	if parameters.ConnectionPool == nil {
		return nil, errors.New("must provide connection pool")
	}

	u := sqlUnit{
		unit:           newUnit(parameters.UnitParameters),
		connectionPool: parameters.ConnectionPool,
	}
	return &u, nil
}

// Save commits the new additions, modifications, and removals
// within the work unit to an SQL store.
func (u *sqlUnit) Save() (err error) {

	//start transaction.
	tx, err := u.connectionPool.Begin()
	if err != nil {
		u.logError(err.Error())
		return
	}

	//rollback if there is a panic.
	defer func() {
		if r := recover(); r != nil {
			err = multierr.Combine(
				fmt.Errorf("panic: unable to save work unit\n%v", r), tx.Rollback())
			u.logError("panic: unable to save work unit",
				zap.String("panic", fmt.Sprintf("%v", r)))
		}
	}()

	//insert newly added entities.
	u.logDebug("attempting to insert entities", zap.Int("count", u.additionCount))
	for typeName, additions := range u.additions {
		if err = u.inserters[typeName].Insert(additions...); err != nil {
			err = multierr.Combine(err, tx.Rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}

	//update altered entities.
	u.logDebug("attempting to update entities", zap.Int("count", u.alterationCount))
	for typeName, alterations := range u.alterations {
		if err = u.updaters[typeName].Update(alterations...); err != nil {
			err = multierr.Combine(err, tx.Rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}

	//delete removed entities.
	u.logDebug("attempting to remove entities", zap.Int("count", u.removalCount))
	for typeName, removals := range u.removals {
		if err = u.deleters[typeName].Delete(removals...); err != nil {
			err = multierr.Combine(err, tx.Rollback())
			u.logError(err.Error(), zap.String("typeName", typeName.String()))
			return
		}
	}

	if err = tx.Commit(); err != nil {
		u.logError(err.Error())
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
