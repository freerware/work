<p align="center"><img src="https://user-images.githubusercontent.com/5921929/73911149-1dad9280-4866-11ea-8818-fed1cd49e8b1.png" width="360"></p>

# work
> A compact library for tracking and committing atomic changes to your entities.

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci]
[![Coverage Status][coverage-img]][coverage] [![Release][release-img]][release]
[![License][license-img]][license] [![Blog][blog-img]][blog]

## How to use it?

### Construction

Starting with entities `Foo` and `Bar`,
```go
// entities.
f, b := Foo{}, Bar{}

// type names.
ft, bt := unit.TypeNameOf(f), unit.TypeNameOf(b)

// data mappers.
m := map[unit.TypeName]unit.DataMapper { ft: fdm, bt: fdm }

// 🎉
opts = []unit.Option{ unit.DB(db), unit.DataMappers(m) }
unit, err := unit.New(opts...)
```

### Adding
When creating new entities, use [`Add`][unit-doc]:
```go
additions := []interface{}{ f, b }
err := u.Add(additions...)
```

### Updating
When modifying existing entities, use [`Alter`][unit-doc]:
```go
updates := []interface{}{ f, b }
err := u.Alter(updates...)
```

### Removing
When removing existing entities, use [`Remove`][unit-doc]:
```go
removals := []interface{}{ f, b }
err := u.Remove(removals...)
```

### Registering 
When retrieving existing entities, track their intial state using
[`Register`][unit-doc]:
```go
fetched := []interface{}{ f, b }
err := u.Register(fetched...)
```

### Saving
When you are ready to commit your work unit, use [`Save`][unit-doc]:
```go
ctx := context.Background()
err := u.Save(ctx)
```

### Logging
We use [`zap`][zap] as our logging library of choice. To leverage the logs
emitted from the work units, utilize the [`unit.Logger`][unit-logger-doc]
option with an instance of [`*zap.Logger`][logger-doc] upon creation:
```go
// create logger.
l, _ := zap.NewDevelopment()

opts = []unit.Option{
	unit.DB(db),
	unit.DataMappers(m),
	unit.Logger(l), // 🎉
}
u, err := unit.New(opts...)
```

### Metrics
For emitting metrics, we use [`tally`][tally]. To utilize the metrics emitted
from the work units, leverage the [`unit.Scope`][unit-scope-doc] option
with a [`tally.Scope`][scope-doc] upon creation. Assuming we have a
scope `s`, it would look like so:
```go
opts = []unit.Option{
	unit.DB(db),
	unit.DataMappers(m),
	unit.Scope(s), // 🎉
}
u, err := unit.New(opts...)
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
[`unit.Uniter`][uniter-doc] to create instances of [`unit.`][unit-doc],
like so:
```go
opts = []unit.Option{
	unit.DB(db),
	unit.DataMappers(m),
	unit.Logger(l),
}
uniter := unit.NewUniter(opts...)

// create the unit.
u, err := uniter.Unit()
```

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
