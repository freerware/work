<p align="center"><img src="https://user-images.githubusercontent.com/5921929/73911149-1dad9280-4866-11ea-8818-fed1cd49e8b1.png" width="360"></p>

# work
> A compact library for tracking and committing atomic changes to your entities.

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci]
[![Coverage Status][coverage-img]][coverage] [![Release][release-img]][release]
[![License][license-img]][license] [![Blog][blog-img]][blog]

## What is it?

`work` does the heavy lifting of tracking changes that your application makes
to entities within a particular operation. This is accomplished by using what we
refer to as a "work unit", which is essentially an implementation of the
[Unit Of Work][uow] pattern popularized by Martin Fowler. With work units,
you no longer need to write any code to track, apply, or rollback changes
atomically in your application. This lets you focus on just writing the code
that handles changes when they happen.

## Why use it?

There are a bundle of benefits you get by using work units:

- easier management of changes to your entities.
- rollback of changes when chaos ensues.
- centralization of save and rollback functionality.
- reduced overhead when applying changes.
- decoupling of code triggering changes from code that persists the changes.
- shorter transactions for SQL datastores.

## How to use it?

The following assumes your application has a variable (`dm`) of a type that
satisfies [`work.DataMapper`][data-mapper-doc], and a variable (`db`) of type
[`*sql.DB`][db-doc].

### Construction

Starting with entities `Foo` and `Bar`,
```go
// entities.
f, b := Foo{}, Bar{}

// type names.
ft, bt := work.TypeNameOf(f), work.TypeNameOf(b)

// data mappers.
m := map[work.TypeName]work.UnitDataMapper { ft: dm, bt: dm }

// 🎉
opts = []work.UnitOption{ work.UnitDB(db), work.UnitDataMappers(m) }
unit, err := work.NewUnit(opts...)
```

### Adding
When creating new entities, use [`Add`][unit-doc]:
```go
additions := []interface{}{ f, b }
err := unit.Add(additions...)
```

### Updating
When modifying existing entities, use [`Alter`][unit-doc]:
```go
updates := []interface{}{ f, b }
err := unit.Alter(updates...)
```

### Removing
When removing existing entities, use [`Remove`][unit-doc]:
```go
removals := []interface{}{ f, b }
err := unit.Remove(removals...)
```

### Registering 
When retrieving existing entities, track their intial state using
[`Register`][unit-doc]:
```go
fetched := []interface{}{ f, b }
err := unit.Register(fetched...)
```

### Saving
When you are ready to commit your work unit, use [`Save`][unit-doc]:
```go
ctx := context.Background()
err := unit.Save(ctx)
```

### Logging
We use [`zap`][zap] as our logging library of choice. To leverage the logs
emitted from the work units, utilize the [`work.UnitLogger`][unit-logger-doc]
option with an instance of [`*zap.Logger`][logger-doc] upon creation:
```go
// create logger.
l, _ := zap.NewDevelopment()

opts = []work.UnitOption{
	work.UnitDB(db),
	work.UnitDataMappers(m),
	work.UnitLogger(l), // 🎉
}
unit, err := work.NewUnit(opts...)
```

### Metrics
For emitting metrics, we use [`tally`][tally]. To utilize the metrics emitted
from the work units, leverage the [`work.UnitScope`][unit-scope-doc] option
with a [`tally.Scope`][scope-doc] upon creation. Assuming we have a
scope `s`, it would look like so:
```go
unit, err := work.NewBestEffortUnit(mappers, work.UnitScope(s))
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
In most circumstances, an application has many aspects that result in the
creation of a work unit. To tackle that challenge, we recommend using
[`work.Uniter`][uniter-doc] to create instances of [`work.Unit`][unit-doc],
like so:
```go
opts = []work.UnitOption{
	work.UnitDB(db),
	work.UnitDataMappers(m),
	work.UnitLogger(l),
}
uniter := work.NewUniter(opts...)

// create the unit.
unit, err := uniter.Unit()
```

## Dependancy Information

As of [`v3.0.0`][modules-release], the project utilizes [modules][modules-doc].
Prior to `v3.0.0`, the project utilized [`dep`][dep] for dependency management.

In order to transition to modules gracefully, we adhered to the
[best practice recommendations][modules-wiki] authored by the Golang team.

## Contribute

Want to lend us a hand? Check out our guidelines for
[contributing][contributing].

## License

We are rocking an [Apache 2.0 license][apache-license] for this project.

## Code of Conduct

Please check out our [code of conduct][code-of-conduct] to get up to speed
how we do things.

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
[unit-logger-doc]: https://godoc.org/github.com/freerware/work#pkg-variables
[unit-scope-doc]: https://godoc.org/github.com/freerware/work#pkg-variables
[modules-doc]: https://golang.org/doc/go1.11#modules
[modules-wiki]: https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher
[modules-release]: https://github.com/freerware/work/releases/tag/v3.0.0
[dep]: https://golang.github.io/dep/
[contributing]: https://github.com/freerware/work/blob/master/CONTRIBUTING.md
[apache-license]: https://github.com/freerware/work/blob/master/LICENSE.txt
[code-of-conduct]: https://github.com/freerware/work/blob/master/CODE_OF_CONDUCT.md
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
