<p align="center"><img src="https://dwglogo.com/wp-content/uploads/2017/08/muscles-clipart-ghoper.gif" width="360"></p>

# work
> A compact library for tracking and committing atomic changes to your entities.

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][coverage-img]][coverage] [![Release][release-img]][release] [![License][license-img]][license] [![Blog][blog-img]][blog]

## What is it?

`work` does the heavy lifting of tracking changes that your application makes to entities within
a particular operation. This is accomplished using what we refer to as a "work unit", which is essentially
an implementation of the [Unit Of Work][uow] pattern popularized by Martin Fowler.
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

The following assumes your application has types (`fdm`, `bdm`) that satisfy the [`SQLDataMapper`][sql-data-mapper-doc] and [`DataMapper`][data-mapper-doc] interfaces, as well as [`*sql.DB`][db-doc] (`db`).

### Construction

Starting with entities `Foo` and `Bar`,
```go
// type names.
fType, bType :=
	work.TypeNameOf(Foo{}), work.TypeNameOf(Bar{})
```

we can create SQL work units:
```go
mappers := map[work.TypeName]work.SQLDataMapper {
	fType: fdm,
	bType: bdm,
}

unit, err := work.NewSQLUnit(mappers, db)
if err != nil {
	panic(err)
}
```

or we can create "best effort" units:
```go
mappers := map[work.TypeName]work.DataMapper {
	fType: fdm,
	bType: bdm,
}

unit, err := work.NewBestEffortUnit(mappers)
if err != nil {
	panic(err)
}
```

### Adding
When creating new entities, use [`Add`][unit-doc]:
```go
additions := interface{}{Foo{}, Bar{}}
unit.Add(additions...)
```

### Updating
When modifying existing entities, use [`Alter`][unit-doc]:
```go
updates := interface{}{Foo{}, Bar{}}
unit.Alter(updates...)
```

### Removing
When removing existing entities, use [`Remove`][unit-doc]:
```go
removals := interface{}{Foo{}, Bar{}}
unit.Remove(removals...)
```

### Registering 
When retrieving existing entities, track their intial state using [`Register`][unit-doc]:
```go
fetched := interface{}{Foo{}, Bar{}}
unit.Register(fetched...)
```

### Saving
When you are ready to commit your work unit, use [`Save`][unit-doc]:
```go
if err := unit.Save(); err != nil {
	panic(err)
}
```

### Logging
We use [`zap`][zap] as our logging library of choice. To leverage the logs emitted from the work units, simply pass in an instance of [`*zap.Logger`][logger-doc] upon creation:
```go
l, _ := zap.NewDevelopment()

// create an SQL unit with logging.
unit, err := work.NewSQLUnit(mappers, db, work.UnitLogger(l))
if err != nil {
	panic(err)
}
```

### Metrics
For emitting metrics, we use [`tally`][tally]. To utilize the metrics emitted from the work units, pass in a [`Scope`][scope-doc] upon creation. Assuming we have an a scope `s`, it would look like so:
```go
unit, err := work.NewBestEffortUnit(mappers, work.UnitScope(s))
if err != nil {
	panic(err)
}
```

#### Emitted Metrics

| Name                             | Type    | Description                                      |
| -------------------------------- | ------- | ------------------------------------------------ |
| [_PREFIX._]unit.save.success     | counter | The number of successful work unit saves.        |
| [_PREFIX._]unit.save             | timer   | The time duration when saving a work unit.       |
| [_PREFIX._]unit.rollback.success | counter | The number of successful work unit rollbacks.    |
| [_PREFIX._]unit.rollback.failure | counter | The number of unsuccessful work unit rollbacks.  |
| [_PREFIX._]unit.rollback         | timer   | The time duration when rolling back a work unit. |

### Uniters
In most circumstances, an application has many aspects that result in the creation of a work unit. To tackle that challenge, we recommend using [`Uniter`][uniter-doc]s to create instances of [`Unit`][unit-doc], like so:
```go
uniter := work.NewSQLUniter(mappers, db)

// create the unit.
unit, err := uniter.Unit()
if err != nil {
	panic(err)
}
```

## Contribute

Want to lend us a hand? Check out our guidelines for [contributing][contributing].

## License

We are rocking an [Apache 2.0 license][apache-license] for this project.

## Code of Conduct

Please check out our [code of conduct][code-of-conduct] to get up to speed how we do things.

## Artwork

Discovered via the interwebs, the artwork was created by Marcus Olsson and Jon Calhoun for [Gophercises][gophercises].

[uow]: https://martinfowler.com/eaaCatalog/unitOfWork.html
[sql-data-mapper-doc]: https://godoc.org/github.com/freerware/work#SQLDataMapper
[data-mapper-doc]: https://godoc.org/github.com/freerware/work#DataMapper
[db-doc]: https://golang.org/pkg/database/sql/#DB
[unit-doc]: https://godoc.org/github.com/freerware/work#Unit
[zap]: https://github.com/uber-go/zap
[tally]: https://github.com/uber-go/tally
[logger-doc]: https://godoc.org/go.uber.org/zap#Logger
[scope-doc]: https://godoc.org/github.com/uber-go/tally#Scope
[uniter-doc]: https://godoc.org/github.com/freerware/work#Uniter
[contributing]: https://github.com/freerware/work/blob/master/CONTRIBUTING.md
[apache-license]: https://github.com/freerware/work/blob/master/LICENSE.txt
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
[blog]: https://medium.com/@freerjm/work-units-ec2da48cf574
[blog-img]: https://img.shields.io/badge/blog-medium-lightgrey
