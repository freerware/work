package work

//Uniter represents a factory for work units.
type Uniter interface {

	//Unit constructs a new work unit.
	Unit() (Unit, error)
}
