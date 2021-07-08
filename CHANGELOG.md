# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.0] Ganymed - 2021-07-09 [:warning:]
Virtual channels. And some other additions.

### Added :sparkles:
- **Virtual Channels** ([#83](https://github.com/hyperledger-labs/go-perun/pull/83), [#114](https://github.com/hyperledger-labs/go-perun/pull/114), [#119](https://github.com/hyperledger-labs/go-perun/pull/119), [#123](https://github.com/hyperledger-labs/go-perun/pull/123)): Go-perun now supports virtual channels. A virtual channel is a channel that is funded and settled completely off-chain and therefore does not incur any on-chain transaction fees.
- **Generic event subscription** ([#36](https://github.com/hyperledger-labs/go-perun/pull/36), [#86](https://github.com/hyperledger-labs/go-perun/pull/86), [#89](https://github.com/hyperledger-labs/go-perun/pull/89), [#94](https://github.com/hyperledger-labs/go-perun/pull/94)): In preparation for implementing a reorg-resistant event subscription, we implemented a generic event subscription that can be used across the whole library.
- **Tutorial announcement** ([c8cff7cb](https://github.com/perun-network/go-perun/commit/c8cff7cbbe97e5b2127e7109483026d1b3938453)): We now have a developer tutorial at [http://tutorial.perun.network](http://tutorial.perun.network).
- **Named errors** ([429a8934](https://github.com/perun-network/go-perun/commit/429a8934666db35d99186d874c23f1102e78750d), [#10](https://github.com/hyperledger-labs/go-perun/pull/10), [#11](https://github.com/hyperledger-labs/go-perun/pull/11), [#26](https://github.com/hyperledger-labs/go-perun/pull/26), [#34](https://github.com/hyperledger-labs/go-perun/pull/34), [#80](https://github.com/hyperledger-labs/go-perun/pull/80)): Specific error types help the library user to identify the cause of an error.
- **Register asset at runtime** [#124](https://github.com/hyperledger-labs/go-perun/pull/124): It is now possible to add assets to the Eth funder at runtime.
- `Gatherer.OnFail` ([e3729a6a](https://github.com/perun-network/go-perun/commit/e3729a6a2f0231273d1cb4ee05ef113c335cdb05)),
Test wallet ([12c78d33](https://github.com/perun-network/go-perun/commit/12c78d33ba339a328d8e5a7b7cf241d9cafca157)),
GitHub CI ([#3](https://github.com/hyperledger-labs/go-perun/pull/3)),
`ConcurrentT.WaitCtx` ([#112](https://github.com/hyperledger-labs/go-perun/pull/112)),
Eth sim backend auto mining ([#104](https://github.com/hyperledger-labs/go-perun/pull/104))

### Changed :construction:
- :boom: **Asset holder validation** ([#111](https://github.com/hyperledger-labs/go-perun/pull/111)): Asset holder validation does no longer include adjudicator validation.
- **Current state in HandleUpdate** [#33](https://github.com/hyperledger-labs/go-perun/pull/33): The update handler now receives the current channel state as a parameter.
- **Funder usability** [#74](https://github.com/hyperledger-labs/go-perun/pull/74): Streamlined Eth funder setup.
- 2021 updates ([2212847d](https://github.com/perun-network/go-perun/commit/2212847de68d683865427c7e11abd48b589f90ee)),
Update generate script ([3f81e47c](https://github.com/perun-network/go-perun/commit/3f81e47cfe436ec42ed4ae9d91742e58f64a0013)),
Update links to HLL ([#6](https://github.com/hyperledger-labs/go-perun/pull/6)),
Update security disclaimer ([#14](https://github.com/hyperledger-labs/go-perun/pull/14), [#51](https://github.com/hyperledger-labs/go-perun/pull/51)),
Delete gitlab templates ([#24](https://github.com/hyperledger-labs/go-perun/pull/24)),
Dependency update ([#30](https://github.com/hyperledger-labs/go-perun/pull/30)),
Document parameters of `NewLedgerChannelProposal` ([#43](https://github.com/hyperledger-labs/go-perun/pull/43)),
CI speed-up ([#44](https://github.com/hyperledger-labs/go-perun/pull/44)),
Refactor Eth channel errors ([#88](https://github.com/hyperledger-labs/go-perun/pull/88)),
Log message type ([#96](https://github.com/hyperledger-labs/go-perun/pull/96))

### Fixed :bug:
- **Cache first channel update** ([#4](https://github.com/hyperledger-labs/go-perun/pull/4), [#129](https://github.com/hyperledger-labs/go-perun/pull/129)): Fixes a bug where a client receives channel messages before completing the channel setup.
- **Subchannel off-chain settlement** [#59](https://github.com/hyperledger-labs/go-perun/pull/59): Sub-channels had to be disputed on-chain before they could be settled. Sub-channels can now be collaboratively settled off-chain.
- **ERC20 depositor nonce mismatch** [#134](https://github.com/hyperledger-labs/go-perun/pull/134): Fixes an issue where the ERC20 depositor sometimes was not incrementing the transaction nonce correctly.

- Unitialized funder variable ([af207adb](https://github.com/perun-network/go-perun/commit/af207adb385329f2b5c0af0fff90c495639a7bf5)),
EndpointRegistry retry and timeout ([28e535bb](https://github.com/perun-network/go-perun/commit/28e535bb6302959c321876968bf1083473824675)),
Watcher return ([cf9279c9](https://github.com/perun-network/go-perun/commit/cf9279c99f73db1f665c072edcd82167888fe83f)),
Thread-safe test wallet [#17](https://github.com/hyperledger-labs/go-perun/pull/17),
Withdraw variable capture [#50](https://github.com/hyperledger-labs/go-perun/pull/50),
`NewRandomLedgerChannelProposal` consistency [#55](https://github.com/hyperledger-labs/go-perun/pull/55),
Enable logging per default in package client [#66](https://github.com/hyperledger-labs/go-perun/pull/66),
Stabilize `BlockTimeout` test [#90](https://github.com/hyperledger-labs/go-perun/pull/90),
Fix state hash test [#120](https://github.com/hyperledger-labs/go-perun/pull/120),
Ensure custom error progagation [#126](https://github.com/hyperledger-labs/go-perun/pull/126)

### Security :lock:
- **Ensure correct params ID after deserialization** [#60](https://github.com/hyperledger-labs/go-perun/pull/60): Parameter deserialization did not assert that the encoded channel ID is correct. This is now fixed.
- **Signature verification for sub-channel funding and settlement** [#61](https://github.com/hyperledger-labs/go-perun/pull/61): Sub-channel funding and settlement requires an automated update in the parent channel. The signatures on that automated update were not correctly verified. This is now fixed.

## [0.6.0] Fenrir - 2020-12-18 [:warning:]
Support for on-chain progression of app channels.

### Added
- On-chain progression: The channel watcher is now interactive and informs the client when a channel has been registered on-chain. If the channel has a defined app logic with a valid state transition logic, clients can individually progress the app state on-chain according to the defined state transition logic by calling `ProgressBy` on the channel object.
- Settle with sub-channel disputes: Ledger channels can now be settled with funds locked in disputed sub-channels if the corresponding sub-channels have been registered and the disputes have been resolved on-chain.

### Changed
- The channel watcher logic changed. The channel watcher now takes as input an event handler which gets notified about on-chain channel events. Before, the watcher automatically settled a channel in case of a dispute. Now, the watcher will automatically detect if an old state has been registered, refute with the most recent one, and notify the user. If the channel has a defined application logic, the user can further progress the channel on-chain. It is within the responsibility of the framework user to finally settle the channel and withdraw the funds.
- The channel settling logic changed. Before, a call to `Settle` on a channel object automatically registered the channel, concluded it, and withdrew the funds. Now, to accomodate on-chain progression functionality, the user must call `Register` independetly before being able to settle the channel. Afterwards, for app channels, the user has the opportunity to  progress the channel state on-chain by calling `ProgressBy`. Finally, the user can settle the channel by calling `Settle`.
- `ContractBackend.NewTransactor` now sets the context on `TransactOpts`. Furthermore, parameter `value` has been removed.

### Fixed
- Persistence: Sub-channels are now persisted and restored properly.

## [0.5.2] European Ecstasy - 2020-11-05 [:warning:]
ERC20 and Funding Agreement support and many test fixes.

### Added
- Funding agreement support: when proposing channels, the proposer can now
  optionally suggest to reallocate the funding responsibilities among the
  participants so that it differs from the initial channel allocation.
- ERC20 support.
- A context-aware `WaitGroup` implementation to `pkg/sync`.

### Changed
- The `eth/wallet/hd.Transactor` to use the correct `types.Signer` if the
  `Wallet` has a correct `SignHash` method.
- The Go version to 1.15.

### Fixed
- Many timeouts in tests that made slow CI pipelines fail.
- A bug in the Concurrent Tester.
- The `Client` role tests to not deadlock so easily.

## [0.5.0] Europa - 2020-10-17 [:warning:]
The Sub-Channels release, enabling fully generalized state channels.

### Added
- Sub-Channels: run app-channels inside parent (ledger) channels.
- Special `NoApp` for channels without app. Skips force-execution phase in
  disputes.
- Optimized `Channel.SettleSecondary` settlement method for the responding
  settler during optimistic channel settlements. Avoids wasting gas by not
  sending unnecessary transactions.
- `ErrorGatherer` type to `pkg/errors` package for errors accumulation.
- Transactor abstraction to allow different wallet implementations for
  transaction sending in Ethereum backend.
- App registry so that multiple apps can be used in a single program instance.

### Changed
- `channel.Update` to accept a `channel.State` instead of `channel.ChannelUpdate`.
  This simplifies the usage.
- Contracts updated to handle sub-channels.
- Contracts now have distinct dispute and force-execution phases.
- Channel proposal protocol now uses shared nonce from all channel peers.

### Fixed
- Channel peers persistence in key-value persistence backend.

## [0.4.0] Despina - 2020-07-23 [:warning:]
Introduced a wire messaging abstraction. License changed to Apache 2.0.

### Added
- Wire messages are now sent and received over an abstract `wire.Bus`. It serves
  as a wire messaging abstraction for the `client` package using pub/sub
  semantics.
  - `wire.Msg`s are wrapped in `Envelope`s that have a sender and recipient.
  - Two implementations available:
    - `wire.LocalBus` for multiple clients running in the same program instance.
    - `wire/net.Bus` for wire connections over networks.
- Consistent use of `wallet.Address`es as map keys (`wallet.AddrKey`).
- Ordering to `wallet.Addresses`es to resolve ties in protocols.
- Contract validation to Ethereum backend.
- Consistent creation of PRNGs in tests (`pkg/test.Prng`).

### Changed
- License to Apache 2.0.
- The packages `peer`, `wire` and `pkg/io` were restructured:
  - Serialization code was moved from `wire` into `pkg/io`.
  - The `peer` package was merged into the `wire` package.
  - Networking-specific `wire` components were moved into `wire/net`.
  - The simple TCP/IP and Unix `Dialer` and `Listener` implementations were
    moved into `wire/net/simple`.
- The `ProposalHandler` and `UpdateHandler` interfaces' methods were renamed to
  explicitly name what they handle (`HandleProposal` and `HandleUpdate`).
- The keyvalue persister uses an improved data model and doesn't cache
  peer-channels any more.
- `Channel.Peers` now returns the full list of channel network peers, including
  the own `wire.Address`.

### Fixed
- A race in `client` synchronization.

### Removed
- The `net` package, as it didn't contain anything useful.

## [0.3.0] Charon - 2020-05-29 [:warning:]
Added persistence module to persist channel state data and handle client
shutdowns/restarts, as well as disconnects/reconnects.

### Added
- Persistence:
  - Persister, Restorer, ChannelIterator interfaces to allow for multiple
    persistence implementations.
    - sortedkv implementation provided (in-memory and LevelDB).
  - States and signatures are constantly persisted while channels progress.
  - Clients restore all saved channels on startup. State is synchronized with peers.
  - `Client.OnNewChannel` callback registration to deal with restored
    channels.
- Wallet interface for account unlocking abstraction.
  - Used during persistence to restore secret keys for signing.
  - Implemented for the Ethereum and simulated backend.
- Peer disconnect/reconnect handling.
- `Channel.UpdateBy` functional channel update method for better usability.

### Changed
- License changed to Apache 2.0.
- Replaced `Channel.ListenUpdates` and `Client.HandleChannelProposals` with
  `Client.Handle(ProposalHandler, UpdateHandler)` - a single common handler
  routine per client.
- Adapted client to new persistence layer and wallet.
- Made Ethereum interactions idempotent (increased safety).
- Moved subpackage `db` to `pkg/sortedkv`.
- Swapped Balance dimensions of type `channel.Allocation`.
- Random type generators in package `channel/test` now accept options to
  customize random data generation.
- Channels now get automatically closed when peers disconnect (and restored on reconnect).

### Fixed
- Ethereum backend: No funding transactions for zero own initial channel balances.

## [0.2.0] Belinda - 2020-03-23 [:warning:]
Added direct disputes and watcher for two-party ledger channels, much polishing
(refactors, bug fixes, documentation).

### Added
- Ledger state channel disputes and watcher.
- `channel.Adjudicator` interface and Ethereum implementation for registering
  channel states and withdrawing concluded channels.
- [Ethereum contracts](https://github.com/hyperledger-labs/perun-eth-contracts/) for disputes
- [Godoc](https://pkg.go.dev/perun.network/go-perun)
- Changelog
- [goreportcard](https://goreportcard.com/report/github.com/hyperledger-labs/go-perun)
- TCP and unix socket `peer.Dialer` and `Listener` implementations.
- `Eventually` tester in `pkg/test` to repeatedly run tests until they
  succeed.
- Concurrent testing tool in `pkg/test` to be able to call `require` in tests
  with multiple go routines.

### Changed
- `client.New` now needs a `Funder` and `Adjudicator`, instead of a `Settler`.
- `Serializable` renamed to `Serializer`.
- Unified backend imports.
- `pkg/io/test/bytewiseReader` to `iotest.OneByteReader`.
- Improved peer message handling mechanism.
- Consistent handling of `nil` arguments in exported functions.
- Many refactors to improve the overall code quality and documentation.
- Updated Ethereum contract bindings to newest version.

### Removed
- `wallet.Wallet` interface and `sim` backend implementation - it was never used.
- `ethereum` and `sim/wallet.NewAddressFromBytes` - only `wallet.DecodeAddress`
  should be used to create an `Address` from bytes.
- `channel/machine` Phase subscription logic.
- `channel.Settler` interface and backend implementations - replaced by `Adjudicator`.

### Fixed
- Reduced cyclomatic complexity of complex functions.
- Deadlock in two-party payment channel test.
- Ethereum backend test timeouts and instabilities.
- Many minor bug fixes, mainly concurrency issues in tests.

## [0.1.0] Ariel - 2019-12-20 [:warning:]
Initial release.

### Added
- Two-party ledger state channels.
- Cooperatively settling two-party ledger channels.
- Ethereum blockchain backend.
- Logrus logging backend.


## Legend
- :warning: This is a pre-release and not intended for usage with real funds.


[:warning:]: #:warning:

[Unreleased]: https://github.com/hyperledger-labs/go-perun/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/hyperledger-labs/go-perun/compare/v0.5.0...v0.5.2
[0.5.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/hyperledger-labs/go-perun/releases/v0.1.0
