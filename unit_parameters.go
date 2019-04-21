package work

import "go.uber.org/zap"

// UnitParameters represents the collection of
// dependencies and configuration needed for a work unit.
type UnitParameters struct {

	//Inserters indicates the mappings between inserters
	//and the entity types they insert.
	Inserters map[TypeName]Inserter

	//Updates indicates the mappings between updaters
	//and the entity types they update.
	Updaters map[TypeName]Updater

	//Deleters indicates the mappings between deleters
	//and the entity types they delete.
	Deleters map[TypeName]Deleter

	//Logger represents the logger that the work unit will utilize.
	Logger *zap.Logger
}
