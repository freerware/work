package work

import "database/sql"

// SQLUnitParameters represents the dependencies and configuration
// required for SQL work units.
type SQLUnitParameters struct {
	UnitParameters

	ConnectionPool *sql.DB
}
