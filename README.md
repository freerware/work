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

- easier management of changes to your entities.
- automatic rollback of changes when chaos ensues.
- centralization of save and rollback functionality.
- reduced overhead when applying changes.
- decoupling of code triggering changes from code that persists the changes.
- production-ready logs and metrics.
- works with your existing persistence layer.
- automatic and configurable retries.

For SQL datastores, also enjoy:

- one transaction, one connection per unit.
- consolidates persistence operations into three operations, regardless of
  the amount of entity changes.
- shorter transaction times.
  - transaction is opened only once the unit is ready to be saved.
  - transaction only remains open as long as it takes for the unit to be saved.
- proper threading of `context.Context` with `database/sql`.

## Release information

| Version | Supported | Documentation                 |
|---------|-----------|-------------------------------|
| `V4`    |  ✅       | [See][v4-docs] `v4/README.md` |
| `V3`    |  ✅       | None                          |
| `V2`    |  ❌       | None                          |
| `V1`    |  ❌       | None                          |

### V4

#### [4.0.0-beta][v4.0.0-beta.4]

- Introduces the work unit cache.
  - Each time the `Register` method is called, the provided entities will be placed in a cache if deemed eligible (have implemented the `identifierer` or `ider` interfaces).
  - Entities will be removed from the cache if specified to `Alter` or `Remove`.

#### [4.0.0-beta.3][v4.0.0-beta.3]

- Various dependency upgrades to address vulnerability [alerts][dependabot-alerts].
	- Upgraded `github.com/uber-go/tally` dependency to version `v3.4.2`.
	- Upgraded `github.com/stretchr/testify` dependency to version `v1.8.0`.
	- Upgraded `go.uber.org/zap` dependency to version `v1.21.1`.

#### [4.0.0-beta.2][v4.0.0-beta.2]

- Introduce initial round of benchmarks.
- Introduce support for 4 more additional metrics.
  - `unit.retry.attempt`
  - `unit.insert`
  - `unit.update`
  - `unit.delete`
- Improve documentation & switch to pkg.go.dev.
- Introduce metric demo.
  - `make demo`

#### [4.0.0-beta][v4.0.0-beta]

- Introduce `unit` package for aliasing.
  - Reduces API footprint.
  - Often "flows" better.
- Introduce retries and related configuration.
- Reconsolidate data mappers abstractions into single `DataMapper` interface.
- Introduce `MapperContext`.
- Alter `Save` to be `context.Context` aware.
- Refactor `work.NewUnit` to dynamically choose which type of work unit to
  create based on provided options.
- Reconsolidate uniter functionality.

### V3

#### [3.2.1][v3.2.1]

- Various dependency upgrades to address vulnerability [alerts][dependabot-alerts].
	- Upgraded `github.com/uber-go/tally` dependency to version `v3.4.2`.
	- Upgraded `github.com/stretchr/testify` dependency to version `v1.8.0`.
	- Upgraded `go.uber.org/zap` dependency to version `v1.21.1`.

#### [3.2.0][v3.2.0]

- Introduce [lifecycle actions][actions-pr].
- Introduce [concurrency support][concurrency-pr].

#### [3.0.0][v3.0.0]

- Introduce support for Go modules.

### V2

- NO LONGER SUPPORTED. CODE REMOVED. SEE `v2.x.x` [TAGS][tags].

### V1

- NO LONGER SUPPORTED. CODE REMOVED. SEE `v1.x.x` [TAGS][tags].

> Versions `1.x.x` and `2.x.x` are no longer supported. Please upgrade to
`3.x.x+` to receive the latest and greatest features, such as
[lifecycle actions][actions-pr] and [concurrency support][concurrency-pr]!

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
[modules-doc]: https://golang.org/doc/go1.11#modules
[modules-wiki]: https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher
[modules-release]: https://github.com/freerware/work/releases/tag/v3.0.0
[dep]: https://golang.github.io/dep/
[contributing]: https://github.com/freerware/work/blob/master/CONTRIBUTING.md
[apache-license]: https://github.com/freerware/work/blob/master/LICENSE.txt
[code-of-conduct]: https://github.com/freerware/work/blob/master/CODE_OF_CONDUCT.md
[concurrency-pr]: https://github.com/freerware/work/pull/35
[actions-pr]: https://github.com/freerware/work/pull/30
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
[v4-docs]: https://github.com/freerware/work/blob/master/v4/README.md
[v3.2.0]: https://github.com/freerware/work/releases/tag/v3.2.0
[v3.2.1]: https://github.com/freerware/work/releases/tag/v3.2.1
[v3.0.0]: https://github.com/freerware/work/releases/tag/v3.0.0
[v4.0.0-beta]: https://github.com/freerware/work/releases/tag/v4.0.0-beta
[v4.0.0-beta.2]: https://github.com/freerware/work/releases/tag/v4.0.0-beta.2
[v4.0.0-beta.3]: https://github.com/freerware/work/releases/tag/v4.0.0-beta.3
[v4.0.0-beta.4]: https://github.com/freerware/work/releases/tag/v4.0.0-beta.4
[tags]: https://github.com/freerware/work/tags
[dependabot-alerts]: https://github.com/freerware/work/security/dependabot?q=is%3Aclosed
