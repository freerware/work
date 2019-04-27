package work

type sqlUniter struct {
	parameters SQLUnitParameters
}

//NewSQLUniter constructs a new SQL work unit factory.
func NewSQLUniter(parameters SQLUnitParameters) Uniter {
	return &sqlUniter{parameters: parameters}
}

// Unit constructs a new SQL work unit.
func (u *sqlUniter) Unit() (Unit, error) {
	return NewSQLUnit(u.parameters)
}
