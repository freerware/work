# work-it
> A compact library for tracking and committing changes to your entities.

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci]

## Overview

`work-it` defines the core abstractions necessary to orchestrate the execution of what we refer to as
a "work unit". A work unit is essentially anologous to the Unit Of Work pattern, popularized by Martin Fowler.

## Usage

Out of the box, `work-it` provides you generic interfaces that you can implement to achieve creation and execution of
work units in your application. In addition, it includes a complete implementation for work units against SQL stores.

To provide an example, let's start by providing a sample implementation of a type in your code that is responsible
for updating a particular type of entity in the SQL database. Here we will create a Data Mapper (to tip our hats
again to Martin Fowler) to perform the job:

```go

// EntityDataMapper is responsible for issuing the database calls to modify
// the underlying SQL database. Here, we have used an "in-memory database" for
// the sake of example.
type EntityDataMapper struct {
	inMemoryDB EntityDB
}

// NewEntityDataMapper constructs a new EntityDataMapper.
func NewEntityDataMapper(db EntityDB) EntityDataMapper {
	return EntityDataMapper{inMemoryDB: db}
}

// Insert is responsible for inserting all of newly created entities of the
// type that this data mapper is responsible for. The work unit will call
// this method when it had new entities to create.
func (dm *EntityDataMapper) Insert(entities ...interface{}) error {
	e := []Entity{}
	for _, entity := range entities {
		u, ok := entity.(Entity)
		if !ok {
			return errors.New("unrecognized type")
		}
		e = append(e, u)
	}
	return dm._Insert(e)
}

// This here is a more strongly typed Insert method. This method is essentially
// being adapted in the above method.
func (dm *EntityDataMapper) _Insert(entities ...Entity) error {
	for _, entity := range entities {
		if _, ok := dm.inMemoryDB[entity.ID()]; !ok {
			dm.inMemoryDB[entity.ID()] = entity
		}
	}
}

// Update is responsible for updating existing entities of the
// type that this data mapper is responsible for. The work unit will call
// this method when it contains modifications for existing entities.
func (dm *EntityDataMapper) Update(entities ...interface{}) error {
	e := []Entity{}
	for _, entity := range entities {
		u, ok := entity.(Entity)
		if !ok {
			return errors.New("unrecognized type")
		}
		e = append(e, u)
	}
	return dm._Update(e)
}

// This here is a more strongly typed Update method. This method is essentially
// being adapted in the above method.
func (dm *EntityDataMapper) _Update(entities ...Entity) error {
	for _, entity := range entities {
		if _, ok := dm.inMemoryDB[entity.ID()]; !ok {
			dm.inMemoryDB[entity.ID()] = entity
		}
	}
}

// Delete is responsible for removing existing entities of the
// type that this data mapper is responsible for. The work unit will call
// this method when it contains removals of existing entities.
func (dm *EntityDataMapper) Delete(entities ...interface{}) error {
	e := []Entity{}
	for _, entity := range entities {
		u, ok := entity.(Entity)
		if !ok {
			return errors.New("unrecognized type")
		}
		e = append(e, u)
	}
	return dm._Delete(e)
}

// This here is a more strongly typed Delete method. This method is essentially
// being adapted in the above method.
func (dm *EntityDataMapper) _Delete(entities ...Entity) error {
	for _, entity := range entities {
		if _, ok := dm.inMemoryDB[entity.ID()]; !ok {
			dm.inMemoryDB[entity.ID()] = entity
		}
	}
}

// Returns the name of the type that this data mapper works with.
func (dm *EntityDataMapper) Type() work.TypeName {
	return work.TypeNameOf(Entity{})
}

```

Next, elsewhere in our code where we need to begin tracking changes to commit to the database, we
create our work unit:

```go

// our database connection pool (likely an instance of *sql.DB).
pool := ...

...

// create work unit.
inserters := make(map[TypeName]Inserter)
inserters[dm.Type()] = &dm
updaters := make(map[TypeName]Updater)
updaters[dm.Type()] = &dm
deleters := make(map[TypeName]Deleter)
deleters[dm.Type()] = &dm
params := work.SQLWorkUnitParameters {
	ConnectionPool: pool,
	WorkUnitParameters: work.WorkUnitParameters {
		Inserters: inserters,
		Updaters: updaters,
		Deleters: deleters,
	}
}
unit := work.NewSQLWorkUnit(params)
```

And then finally, we indicate to the work unit what changes should be tracked, and then save them!

```go

...

// entity that we are creating.
newEntity := getNewEntity()

// entity that we are removing.
removedEntity := getRemovedEntity()

// entity that we are updating.
updatedEntity := getUpdatedEntity()

//track these changes with the work unit.

unit.Add(newEntity)
unit.Alter(updatedEntity)
unit.Remove(removedEntity)

//commit the changes.
if err := unit.Save(); err != nil {
	panic("unable to commit changes")
}
```
[doc-img]: https://godoc.org/github.com/freerware/work-it?status.svg
[doc]: https://godoc.org/github.com/freerware/work-it
[ci-img]: https://travis-ci.org/freerware/work-it.svg?branch=master
[ci]: https://travis-ci.org/freerware/work-it
