# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0]: Belinda - 2020-02-21 [:warning:]
Added direct disputes for two-party ledger channels, much polishing (refactors, bug fixes, documentation).

### Added
- Ledger state channel disputes
- Ethereum contracts for disputes
- Public [Github wiki](https://github.com/perun-network/go-perun/wiki)
- [Godoc](https://godoc.org/perun.network/go-perun)
- Changelog
- [TravisCI](https://travis-ci.org/perun-network)
- [goreportcard](https://goreportcard.com/report/github.com/perun-network/go-perun)
- [codeclimate](https://codeclimate.com/github/perun-network/go-perun)
  
### Changed
- `Serializable` renamed to `Serializer`
- Unified backend imports
- `pkg/io/test/bytewiseReader` to `iotest.OneByteReader`
- Improved message handling mechanism.
- Many refactors to improve the overall code quality.

### Removed
- Wallet interface.
- ethereum/wallet `NewAddressFromBytes`
- `channel/machine` subscription logic

### Fixed
- Reduced cyclomatic complexity of complex functions.
- Deadlock in two-party payment channel test.
- `TestSettler_MultipleSettles` timeout.
- Many minor bug fixes, mainly concurrency issues in tests.

## [0.1.0]: Ariel - 2019-12-20 [:warning:]
Initial release, intended to receive feedback.

### Added
- Two-party ledger state channels
- Cooperatively settling two-party ledger channels


## Legend
- <a name=":warning:">:warning:</a>: This release is not suited for practical use with real money.


[:warning:]: #:warning:

[Unreleased]: https://github.com/perun-network/go-perun/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/perun-network/go-perun/compare/tag/v0.1.0...v0.2.0
[0.1.0]: https://github.com/perun-network/go-perun/releases/tag/v0.1.0
