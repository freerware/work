package work

//Inserter represents a creator of entities.
type Inserter interface {

	// Insert creates the provided entities in a persistent store.
	Insert(...interface{}) error
}
