# work
> A compact library for tracking and committing atomic changes to your entities.

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][coverage-img]][coverage] [![Release][release-img]][release] [![License][license-img]][license]

## What is it?

`work` does the heavy lifting of tracking changes that your application makes to entities within
a particular operation. This is accomplished using what we refer to as a "work unit", which is essentially
an implementation of the [Unit Of Work](https://martinfowler.com/eaaCatalog/unitOfWork.html) pattern popularized by Martin Fowler.
With work units, you no longer need to write any code to track, apply, or rollback changes atomically in your application.
This lets you focus on just writing the code that handles changes when they happen.

## Why use it?

There are a bundle of benefits you get by using work units:

- easy management of changes to your entities.
- centralization of save and rollback functionality.
- rollback of changes when chaos ensues.
- reduce overhead when applying changes.
- decouple code triggering changes from code that applies the changes.

## Example Usage

One type of work unit is the [`SQLUnit`](https://github.com/freerware/work/blob/master/sql_unit.go). This kind of unit manages orchestrating
changes when using a relational database, such as MySQL or Postgres. 

First, let's kick things off by defining a type that will be responsible for actually 
interfacing with our database:

```golang
// EntityDataMapper is responsible for issuing the database calls to modify
// the underlying SQL database. 
type EntityDataMapper struct {
	...
}

// NewEntityDataMapper constructs a new EntityDataMapper.
func NewEntityDataMapper(db *sql.DB) EntityDataMapper {
	return EntityDataMapper{...}
}

// Insert is responsible for inserting all of newly created entities of the
// type that this data mapper is responsible for. The work unit will call
// this method when it has new entities to create.
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
	// inserts the entities to the SQL database using the `database/sql` package.
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
	// updates the entities in the SQL database using the `database/sql` package.
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
	// deletes the entities from the SQL database using the `database/sql` package.
}

// Returns the name of the type that this data mapper works with.
func (dm *EntityDataMapper) Type() work.TypeName {
	return work.TypeNameOf(Entity{})
}
```

Next, elsewhere in our code where we need to begin tracking changes to commit to the database, we
create our work unit:

```golang
// our database connection pool.
pool, _ := sql.Open(...)

...

// create work unit.
inserters := make(map[TypeName]Inserter)
inserters[dm.Type()] = &dm
updaters := make(map[TypeName]Updater)
updaters[dm.Type()] = &dm
deleters := make(map[TypeName]Deleter)
deleters[dm.Type()] = &dm
params := work.SQLUnitParameters {
	ConnectionPool: pool,
	UnitParameters: work.UnitParameters {
		Inserters: inserters,
		Updaters: updaters,
		Deleters: deleters,
	}
}
unit, err := work.NewSQLUnit(params)
if err != nil {
	panic("unable to create work unit")
}
```

And then finally, we indicate to the work unit what changes should be tracked, and then save them!

```golang

...

// entities that we are creating.
newEntities := //...

// entities that we are removing.
removedEntities := //...

// entities that we are updating.
updatedEntities := //...

//track these changes with the work unit.

unit.Add(newEntities...)
unit.Alter(updatedEntities...)
unit.Remove(removedEntities...)

//commit the changes.
if err := unit.Save(); err != nil {
	panic("unable to commit changes")
}
```

[doc-img]: https://godoc.org/github.com/freerware/work?status.svg
[doc]: https://godoc.org/github.com/freerware/work
[ci-img]: https://travis-ci.org/freerware/work.svg?branch=master
[ci]: https://travis-ci.org/freerware/work
[coverage-img]: https://coveralls.io/repos/github/freerware/work/badge.svg?branch=master
[coverage]: https://coveralls.io/github/freerware/work?branch=master
[license]: https://opensource.org/licenses/Apache-2.0
[license-img]: https://img.shields.io/badge/License-Apache%202.0-blue.svg
[release]: https://github.com/freerware/work/releases
[release-img]: https://img.shields.io/github/tag/freerware/work.svg?label=version
