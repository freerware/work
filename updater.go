package work

//Updater represents an alterer of entities.
type Updater interface {

	// Update modifies the provided entities within a persistent store.
	Update(...interface{}) error
}
