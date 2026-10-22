# Contributing to autobrr

Thanks for taking interest in contribution! We welcome anyone who wants to contribute.

If you have an idea for a bigger feature or a change then we are happy to discuss it before you start working on it.  
It is usually a good idea to make sure it aligns with the project and is a good fit.  
Open an issue or post in #dev-general on [Discord](https://discord.autobrr.com).

This document is a guide to help you through the process of contributing to autobrr.

## Become a contributor

* Code: new features, bug fixes, improvements
* Report bugs
* Documentation: The docs repo can be found here: [github.com/autobrr/autobrr.com](https://github.com/autobrr/autobrr.com)

## Developer guide

This guide helps you get started developing autobrr.

## Dependencies

Make sure you have the following dependencies installed before setting up your developer environment:

- [Git](https://git-scm.com/)
- [Go](https://golang.org/dl/) (see [go.mod](go.mod#L3) for minimum required version)
- macOS with Xcode, to build and run the iOS app

## How to contribute

- **Fork and Clone:** [Fork the autobrr repository](https://github.com/autobrr/autobrr/fork) and clone it to start working on your changes.
- **Branching:** Create a new branch for your changes. Use a descriptive name for easy understanding.
  - Checkout a new branch for your fix or feature `git checkout -b fix/filters-issue`
- **Coding:** Comment non-obvious logic - see [AGENTS.md](AGENTS.md) for code style and comment conventions. With go use `go fmt`
- **Commit Guidelines:** We appreciate the use of [Conventional Commit Guidelines](https://www.conventionalcommits.org/en/v1.0.0/#summary) when writing your commits.
  - Examples: `fix(indexers): Mock improve parsing`, `feat(notifications): add NewService`
  - There is no need for force pushing or rebasing. We squash commits on merge to keep the history clean and manageable.
  - The PR title becomes the squashed commit message, so use the conventional commit format for the title as well.
- **Pull Requests:** Submit a pull request from your Fork with a clear description of your changes. Reference any related issues.
  - Target the `develop` branch.
  - Fill out the pull request template, including the AI disclosure section.
  - Mark it as Draft if it's still in progress.
- **Code Review:** Be open to feedback during the code review process.

## Development environment

The engine is written in Go and the app in Swift with SwiftUI. The Go engine is built into a framework with gomobile, and the app calls it in-process. There is no standalone server.

Clone the project and change dir:

```shell
git clone https://github.com/YOURNAME/autobrr.git && cd autobrr
```

Install the Go dependencies with:

```shell
make deps
```

## iOS app

Building the app needs macOS with Xcode, [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile) and [XcodeGen](https://github.com/yonaskolb/XcodeGen). See [ios/README.md](ios/README.md).

```shell
make build/ios
cd ios && xcodegen generate && open Autobrr.xcodeproj
```

## Tests

All tests run per commit with GitHub Actions.

### Run backend tests

We have a mix of unit and integration tests.

Run all non-integration tests:

```shell
go test -v ./...
```

### Run integration tests

The integration tests run against an in-memory SQLite database:

```shell
go test ./... -tags=integration
```

## Adding a new indexer

Adding support for a new indexer is the most common contribution. Indexer definitions are YAML files in [internal/indexer/definitions](internal/indexer/definitions) - start from an existing definition for a similar tracker and adjust it.

Definitions are embedded into the engine at build time, so rebuild the app after changing one. Use the mock indexer below to test announce parsing end-to-end.

### Retiring an indexer

When a tracker shuts down, remove its active definition and add a minimal tombstone at `internal/indexer/definitions/deprecated/<identifier>.yaml`. The tombstone records its identifier, display name, reason, issue or pull request URL, and deprecation date. It lets existing installations archive the configured indexer without deleting user data and preserves friendly names for historical releases.

Keep tombstones across releases so installations can safely skip versions. If the same identifier becomes active again, remove its tombstone; startup reconciliation will restore archived configurations.

Pull request CI rejects a removed or changed identifier that has neither an active definition nor a matching tombstone, and rejects dropping an existing tombstone unless its identifier is active again.

## Mock indexer

We have a mock indexer you can run locally that features:

* Built in IRC server that can send announces
* Mock indexer for downloads
* RSS feed mock
* Webhook mock for External Filters

See the documentation [here](./test/mockindexer/README.md). Add the `customDefinitions` to the `config.toml` and then run it with:

```shell
go run test/mockindexer/main.go
```

* Restart the app for it to load the new mock.yaml definition (in the iOS simulator, `localhost` reaches your Mac)
* Then add it via Settings -> Indexers -> Add, and select Mock Indexer in the list
* Go to Settings -> IRC and toggle the IRC network `Mock Indexer`
* Add a new Filter or add the indexer to an existing filter
* Open a new tab and navigate to [http://localhost:3999](http://localhost:3999) and put the example announce in the input then hit enter
