<p align="center"><img src="https://user-images.githubusercontent.com/5921929/73911149-1dad9280-4866-11ea-8818-fed1cd49e8b1.png" width="360"></p>

# work
> A compact library for tracking and committing atomic changes to your entities.

[![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci]
[![Coverage Status][coverage-img]][coverage] [![Release][release-img]][release]
[![License][license-img]][license] [![Blog][blog-img]][blog]

## Demo

`make demo` (requires `docker-compose`).

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

<p align="center"><img src="https://user-images.githubusercontent.com/5921929/106403546-191daa80-63e4-11eb-98b5-6b5d1989bacb.gif" width="960"></p>

| Name                             | Type    | Description                                      |
| -------------------------------- | ------- | ------------------------------------------------ |
| [_PREFIX._]unit.save.success     | counter | The number of successful work unit saves.        |
| [_PREFIX._]unit.save             | timer   | The time duration when saving a work unit.       |
| [_PREFIX._]unit.rollback.success | counter | The number of successful work unit rollbacks.    |
| [_PREFIX._]unit.rollback.failure | counter | The number of unsuccessful work unit rollbacks.  |
| [_PREFIX._]unit.rollback         | timer   | The time duration when rolling back a work unit. |
| [_PREFIX._]unit.retry.attempt    | counter | The number of retry attempts.                    |
| [_PREFIX._]unit.insert           | counter | The number of successful inserts performed.      |
| [_PREFIX._]unit.update           | counter | The number of successful updates performed.      |
| [_PREFIX._]unit.delete           | counter | The number of successful deletes performed.      |

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

## Frequently Asked Questions (FAQ)

### Are batch data mapper operations supported?

In short, yes.

A work unit can accommodate an arbitrary number of entity types. When creating
the work unit, you indicate the data mappers that it should use when persisting
the desired state. These data mappers are organized by entity type. As such,
batching occurs for each operation and entity type pair.

For example, assume we have a single work unit and have performed a myriad
of unit operations for entities with either a type of `Foo` or `Bar`. All inserts
for entities of type `Foo` will be [passed][insert-method-ref] to the corresponding data mapper in
one shot via the `Insert` [method][insert-method]. This essentially then relinquishes control to you,
the author of the data mapper, to handle all of those entities to be inserted
in however you see fit. You could choose to insert them all into a relational
database using a single `INSERT` query, or perhaps issue an HTTP request to
an API to create all of those entities. However, inserts for entities of type
`Bar` will be batched separately. In fact, it's most likely the data mapper to handle
inserts for `Foo` and `Bar` are completely different types (and maybe even
completely different data stores).

The same applies for other operations such as updates and deletions. All
supported data mapper operations follow this paradigm.

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
[doc-img]: https://pkg.go.dev/badge/github.com/freerware/work/v4.svg
[doc]: https://pkg.go.dev/github.com/freerware/work/v4
[ci-img]: https://github.com/freerware/work/actions/workflows/ci.yaml/badge.svg?branch=master
[ci]: https://github.com/freerware/work/actions/workflows/ci.yaml
[coverage-img]: https://codecov.io/gh/freerware/work/branch/master/graph/badge.svg?token=W5YH9TPP3C
[coverage]: https://codecov.io/gh/freerware/work
[license]: https://opensource.org/licenses/Apache-2.0
[license-img]: https://img.shields.io/badge/License-Apache%202.0-blue.svg
[release]: https://github.com/freerware/work/releases
[release-img]: https://img.shields.io/github/tag/freerware/work.svg?label=version
[blog]: https://medium.com/@freerjm/work-units-ec2da48cf574
[blog-img]: https://img.shields.io/badge/blog-medium-lightgrey
[insert-method]: https://github.com/freerware/work/blob/v4.0.0-beta.2/v4/data_mapper.go#L22
[insert-method-ref]: https://github.com/freerware/work/blob/v4.0.0-beta.2/v4/best_effort_unit.go#L137
