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

For SQL datastores, also enjoy:

- one transaction, one connection per unit.
- three queries max.
  - additions result in a single `INSERT`.
  - alters result in a single `UPDATE`.
  - removals result in a single `DELETE`.
- shorter transaction times.
  - transaction is opened only once the unit is ready to be saved.
  - transaction only remains open as long as it takes for the unit to be saved.

## Release information

### [4.0.0][]

- Introduce `unit` package to alias all types.
- API revamp. Using `work` has never been easier!

### [3.2.0][v3.2.0]

- Introduce [lifecycle actions][actions-pr].
- Introduce [concurrency support][concurrency-pr].

### [3.0.0][v3.0.0]

- Introduce support for Go modules.

### 2.x.x

- NO LONGER SUPPORTED.

### 1.x.x

- NO LONGER SUPPORTED.

## Dependancy Information

As of [`v3.0.0`][modules-release], the project utilizes [modules][modules-doc].
Prior to `v3.0.0`, the project utilized [`dep`][dep] for dependency management.

In order to transition to modules gracefully, we adhered to the
[best practice recommendations][modules-wiki] authored by the Golang team.

## Release information

Versions `1.x.x` and `2.x.x` are currently in maintenance mode. Please upgrade to `3.x.x` to
receive the latest and greatest features, such as [lifecycle actions][actions-pr] and 
[concurrency support][concurrency-pr]!

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
[v3.2.0]: https://github.com/freerware/work/releases/tag/v3.2.0
[v3.0.0]: https://github.com/freerware/work/releases/tag/v3.0.0

Versions `1.x.x` and `2.x.x` are no longer supported. Please upgrade to
`3.x.x+` to receive the latest and greatest features, such as
[lifecycle actions][actions-pr] and [concurrency support][concurrency-pr]!
