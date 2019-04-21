package work

// Deleter represents a remover of entities.
type Deleter interface {

	// Delete removes the provided entities from a persistent store.
	Delete(...interface{}) error
}
