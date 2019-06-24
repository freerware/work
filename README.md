<p align="center"><img src="https://dwglogo.com/wp-content/uploads/2017/08/muscles-clipart-ghoper.gif" width="360"></p>

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

- easier management of changes to your entities.
- rollback of changes when chaos ensues.
- centralization of save and rollback functionality.
- reduced overhead when applying changes.
- decoupling of code triggering changes from code that persists the changes.
- shorter transactions for SQL datastores.

## How to use it?
The following assumes your application has types (`fdm`, `bdm`) that satisfy the [`Inserter`][inserter-doc], [`Updater`][updater-doc], 
and [`Deleter`][deleter-doc] interfaces, as well as an instance of [`*sql.DB`][db-doc] (`db`).

### Construction
Starting with a sample setup,
```go
// type names.
fType, bType :=
	work.TypeNameOf(Foo{}), work.TypeNameOf(Bar{})

// parameters.
i, u, d :=
	make(map[work.TypeName]work.Inserter),
	make(map[work.TypeName]work.Updater),
	make(map[work.TypeName]work.Deleter)
i[fType], i[bType] = fdm, bdm
u[fType], u[bType] = fdm, bdm
d[fType], d[bType] = fdm, bdm
```

we can create SQL work units:
```go
// SQL unit construction.
unit, err := work.NewSQLUnit(work.SQLUnitParameters {
	ConnectionPool: db,
	Inserters: i,
	Updaters: u,
	Deleters: d,
})
if err != nil {
	panic(err)
}
```

or we can create "best effort" units:
```go
// best effort unit construction.
unit, err := work.NewBestEffortUnit(work.UnitParameters {
	Inserters: i,
	Updaters: u,
	Deleters: d,
})
if err != nil {
	panic(err)
}
```

### Adding
When creating a new entity, use [`Add`][unit-doc]:
```go
additions := interface{}{Foo{}, Bar{}}
unit.Add(additions...}
```

### Updating
When modifying an existing entity, use [`Alter`][unit-doc]:
```go
updates := interface{}{Foo{}, Bar{}}
unit.Alter(updates...)
```

### Removing
When removing an existing entity, use [`Remove`][unit-doc]:
```go
removals := interface{}{Foo{}, Bar{}}
unit.Remove(removals...)
```

### Registering 
When retrieving an existing entity, track it's intial state using [`Register`][unit-doc]:
```go
fetched := interface{}{Foo{}, Bar{}}
unit.Register(fetched...}
```

### Saving
When you are ready to commit your work unit, use [`Save`][unit-doc]:
```go
if err := unit.Save(); err != nil {
  panic(err)
}
```

## Contribute

Want to lend us a hand? Check out our guidelines for [contributing][contributing].

## License

We are rocking an [MIT license][mit-license] for this project.

## Code of Conduct

Please check out our [code of conduct][code-of-conduct] to get up to speed how we do things.

## Artwork

Discovered via the interwebs, the artwork was created by Marcus Olsson and Jon Calhoun for [Gophercises][gophercises].

[inserter-doc]: https://godoc.org/github.com/freerware/work#Inserter
[updater-doc]: https://godoc.org/github.com/freerware/work#Updater
[deleter-doc]: https://godoc.org/github.com/freerware/work#Deleter
[db-doc]: https://golang.org/pkg/database/sql/#DB
[unit-doc]: https://godoc.org/github.com/freerware/work#Unit
[contributing]: https://github.com/freerware/work/blob/master/CONTRIBUTING.md
[mit-license]: https://github.com/freerware/work/blob/master/LICENSE.txt
[code-of-conduct]: https://github.com/freerware/work/blob/master/CODE_OF_CONDUCT.md
[gophercises]: https://gophercises.com
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
