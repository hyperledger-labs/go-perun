# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.14.0] Narvi - 2025-07-29 [:boom:]
Added [Libp2p](https://libp2p.io/) wire for go-perun. This enables seamless and secure P2P connection between clients.

### Added :boom:

* Wire implementation with libp2p [#420]

### Fixed
* Removed test backend signature from `authMsg` of exchange address protocol [#422] 

### Changed

* Updade action cache [#420]

* Update workflow go version to 1.23 [#420]

* Update workflow go linter to 2.1 [#420]

[#420]: https://github.com/hyperledger-labs/go-perun/pull/420
[#422]: https://github.com/hyperledger-labs/go-perun/pull/422


## [0.13.0] Metis - 2025-01-19 [:boom:]
Support for multiple backends, allowing multiple address implementations per client. This enables the simultaneous use of several smaller backends, enhancing modularization.

### Added :boom:

* Backend field in Allocation [#410]

* Added interface restrictions to ensure cross-contract compatibility, including new functions and fields in interfaces such as Asset and Address [#410]

### Changed

* Updade action cache [#409]

* Update workflow go version to 1.18 [#410]

* Global Backend map in wire and wallet module [#410]

* Global Randomizer map in wallet and channel tests [#410]

* Participant map to allow multiple addresses per participant [#410] :boom:

* Code refactoring from channel ID map to singular channel ID [#413]

[#409]: https://github.com/hyperledger-labs/go-perun/pull/409
[#410]: https://github.com/hyperledger-labs/go-perun/pull/410
[#413]: https://github.com/hyperledger-labs/go-perun/pull/413

### Legend
- <span id="breaking">:boom:</span> This is a breaking change, e.g., it changes the external API.

[:boom:]: #breaking

## [0.12.0] Leda - 2024-11-19 [:boom:]
Flexibility in funding for payment channels and basic Layer-2 security.

### Added :boom:

* Egoistic funding allows users to wait for their peers to fund before they fund themselves. This change has to be adopted by the Perun backends to be usable ([in case of Ethereum](https://github.com/hyperledger-labs/perun-eth-backend/pull/45)): [#397 ]

* Wire authentication for Layer2 communication between Perun clients, using TLS: [#402] :boom:

* Support for Stellar backend in README: [#408]

### Changed
* Update go to 1.22: [#406 ]

[#397]: https://github.com/hyperledger-labs/go-perun/pull/397
[#406]: https://github.com/hyperledger-labs/go-perun/pull/406
[#402]: https://github.com/hyperledger-labs/go-perun/pull/402
[#408]: https://github.com/hyperledger-labs/go-perun/pull/408

### Legend
- <span id="breaking">:boom:</span> This is a breaking change, e.g., it changes the external API.

[:boom:]: #breaking

## [0.11.0] Kiviuq - 2024-02-21 [:boom:]
Exposure of protobuf converters & `SignedState`, abstraction of tests and bug fixes.

### Added
- Add Fabric to backend list in README: [#377]
- Create new type `TransparentChannel` to expose `SignedState`: [#389]
- Update backend compatibility list in README: [#392]
- Add MAINTAINERS.md file, Update NOTICE: [#394]

### Fixed 
- Fix sub-channel test: [#359]
- Fix Multi-Adjudicator Subscription: [#366]
- Use correct identity for client tests: [#376]
- Fix link to white paper in README: [#379]
- Fix linter copyright year checking in CI: [#389]
- Fix failing unit tests: [#399]

### Changed [:boom:]
- Abstract multiledger test, making it usable by backends: [#355]
- Abstract fund recovery test, making it usable by backends: [#370]
- Abstract virtual channel test, making it usable by backends: [#375]
- Expose protobuf converters: [#384] [#393]
- Use absolute module path in wire.proto: [#383]
- Create AppID Type to generalize app identifiers: [#378] [:boom:]


[#359]: https://github.com/hyperledger-labs/go-perun/pull/359
[#355]: https://github.com/hyperledger-labs/go-perun/pull/355
[#366]: https://github.com/hyperledger-labs/go-perun/pull/366
[#370]: https://github.com/hyperledger-labs/go-perun/pull/370
[#375]: https://github.com/hyperledger-labs/go-perun/pull/375
[#376]: https://github.com/hyperledger-labs/go-perun/pull/376
[#377]: https://github.com/hyperledger-labs/go-perun/pull/377
[#378]: https://github.com/hyperledger-labs/go-perun/pull/378
[#379]: https://github.com/hyperledger-labs/go-perun/pull/379
[#383]: https://github.com/hyperledger-labs/go-perun/pull/383
[#384]: https://github.com/hyperledger-labs/go-perun/pull/384
[#389]: https://github.com/hyperledger-labs/go-perun/pull/389
[#392]: https://github.com/hyperledger-labs/go-perun/pull/392
[#393]: https://github.com/hyperledger-labs/go-perun/pull/393
[#394]: https://github.com/hyperledger-labs/go-perun/pull/394
[#399]: https://github.com/hyperledger-labs/go-perun/pull/399


## [0.10.0] Janus - 2022-05-25 [:warning:]
Multi-ledger payment channels.

### Added
- Multi-ledger payment channels: [#337]
- App channel test: [#339]

### Changed
- Revise dispute test: [#340]
- Enable require in client tests: [#341]

### Fixed
- Satisfy linter warnings: [#342]

### Removed
- Remove go-ethereum dependency: [#338]

[#337]: https://github.com/hyperledger-labs/go-perun/pull/337
[#338]: https://github.com/hyperledger-labs/go-perun/pull/338
[#339]: https://github.com/hyperledger-labs/go-perun/pull/339
[#340]: https://github.com/hyperledger-labs/go-perun/pull/340
[#341]: https://github.com/hyperledger-labs/go-perun/pull/341
[#342]: https://github.com/hyperledger-labs/go-perun/pull/342

## [0.9.1] Io Pioneer - 2022-04-14 [:warning:]
Moved Ethereum backend to separate repository.

### Added
- `CloneAddresses` helper: [#331].
- Wire hybrid bus: [#327].

### Removed
- Moved Ethereum backend to separate repository: [#355].

[#327]: https://github.com/hyperledger-labs/go-perun/pull/327
[#331]: https://github.com/hyperledger-labs/go-perun/pull/331
[#355]: https://github.com/hyperledger-labs/go-perun/pull/335

## [0.9.0] Io - 2022-02-22 [:warning:]
Injectable wire encoding and a `protobuf` wire encoder.

### Added

- :sparkles: **Injectable wire encoding [#233]:** The encoding used for messages that are sent across the network is now injectable.
  - Add binary marshalling: [#272], [#284], [#298].
  - Add serializer interface: [#297], [#325].
  - Change the mechanism for generating `ProposalID`: [#300], [#307].
  - Rename message types [#305].
  - Extend and refactor message serialization tests: [#316], [#317].
- :sparkles: **Protobuf wire encoder [#311]:** A wire encoder based on `protobuf` has been added [#318].
- Minor additions: Introduce `Asset.Equal` [#279], export proposer index [#313].

### Changed

- :boom: Rename `Address.Equals` to `Address.Equal` [#264].
- :boom: Revise `Channel.Update` and `Channel.ForceUpdate`: [#289], [#291], [#306].
- :truck: `pkg/io` migration: [#256], [#271], [#285], [#287].
- :memo: Update README to include references to existing backends [#314].
- :children_crossing: Minor usability improvements: [#268], [#278], [#315].
- :construction_worker: CI updates: [#260], [#274], [#276], [#277].
- :arrow_up: Use `LatestSigner` in Ethereum contract backend and don't set `GasLimit` by default, so that [EIP1559](https://eips.ethereum.org/EIPS/eip-1559) TXs are sent [#322].

### Fixed
- :bug: Fix a bug causing `Channel.Watch` to not work correctly for sub-channels and virtual channels [#251].
- :bug: Fix bugs causing `ContractBackend.confirmNTimes` to block indefintely or fail in some rare cases: [#254], [#309].
- :white_check_mark: Improve test stability: [#310], [#319].

### Security
- :lock: It is now checked that assets are not changed during a state update, which could otherwise cause a channel to become unfunded [#304].

[#233]: https://github.com/hyperledger-labs/go-perun/issues/233
[#311]: https://github.com/hyperledger-labs/go-perun/issues/311

[#251]: https://github.com/hyperledger-labs/go-perun/pull/251
[#254]: https://github.com/hyperledger-labs/go-perun/pull/254
[#256]: https://github.com/hyperledger-labs/go-perun/pull/256
[#260]: https://github.com/hyperledger-labs/go-perun/pull/260
[#264]: https://github.com/hyperledger-labs/go-perun/pull/264
[#268]: https://github.com/hyperledger-labs/go-perun/pull/268
[#271]: https://github.com/hyperledger-labs/go-perun/pull/271
[#272]: https://github.com/hyperledger-labs/go-perun/pull/272
[#274]: https://github.com/hyperledger-labs/go-perun/pull/274
[#276]: https://github.com/hyperledger-labs/go-perun/pull/276
[#277]: https://github.com/hyperledger-labs/go-perun/pull/277
[#278]: https://github.com/hyperledger-labs/go-perun/pull/278
[#279]: https://github.com/hyperledger-labs/go-perun/pull/279
[#284]: https://github.com/hyperledger-labs/go-perun/pull/284
[#285]: https://github.com/hyperledger-labs/go-perun/pull/285
[#287]: https://github.com/hyperledger-labs/go-perun/pull/287
[#289]: https://github.com/hyperledger-labs/go-perun/pull/289
[#291]: https://github.com/hyperledger-labs/go-perun/pull/291
[#297]: https://github.com/hyperledger-labs/go-perun/pull/297
[#298]: https://github.com/hyperledger-labs/go-perun/pull/298
[#300]: https://github.com/hyperledger-labs/go-perun/pull/300
[#304]: https://github.com/hyperledger-labs/go-perun/pull/304
[#305]: https://github.com/hyperledger-labs/go-perun/pull/305
[#306]: https://github.com/hyperledger-labs/go-perun/pull/306
[#307]: https://github.com/hyperledger-labs/go-perun/pull/307
[#309]: https://github.com/hyperledger-labs/go-perun/pull/309
[#310]: https://github.com/hyperledger-labs/go-perun/pull/310
[#313]: https://github.com/hyperledger-labs/go-perun/pull/313
[#314]: https://github.com/hyperledger-labs/go-perun/pull/314
[#315]: https://github.com/hyperledger-labs/go-perun/pull/315
[#316]: https://github.com/hyperledger-labs/go-perun/pull/316
[#317]: https://github.com/hyperledger-labs/go-perun/pull/317
[#318]: https://github.com/hyperledger-labs/go-perun/pull/318
[#319]: https://github.com/hyperledger-labs/go-perun/pull/319
[#322]: https://github.com/hyperledger-labs/go-perun/pull/322
[#325]: https://github.com/hyperledger-labs/go-perun/pull/325


## [0.8.0] Hyperion - 2021-11-08 [:warning:]
Reorg-resistance for the Ethereum backend and support for external Watchtowers.

### Added

- :sparkles: **Ethereum backend: Reorg resistance** [#19](https://github.com/hyperledger-labs/go-perun/issues/19): The Ethereum backend now lets the user specify after how many blocks a transaction or an event should be considered confirmed.
- :sparkles: **Watcher interface** [#172](https://github.com/hyperledger-labs/go-perun/issues/172): The watcher logic is now injectable. The adjudicator takes a watcher instance as a setup parameter. This enables using remote watcher services. A local watcher implementation is provided.

### Changed

- :arrow_up: Update go to v1.17 and go-ethereum to v1.10.12.
- :white_check_mark: Extend support for additional blockchain backends by revising the generic tests (e.g., [#225], [#227], [#228]).
- :children_crossing: Improve usability (e.g., [#124], [#144], [#196], [#204], [#240]).

### Fixed
- :bug: Improve stability (e.g., [#129], [#134], [#148], [#191], [#207], [#218]).

[#225]: https://github.com/hyperledger-labs/go-perun/pull/225
[#227]: https://github.com/hyperledger-labs/go-perun/pull/227
[#228]: https://github.com/hyperledger-labs/go-perun/pull/228
[#124]: https://github.com/hyperledger-labs/go-perun/pull/124
[#144]: https://github.com/hyperledger-labs/go-perun/pull/144
[#204]: https://github.com/hyperledger-labs/go-perun/pull/204
[#129]: https://github.com/hyperledger-labs/go-perun/pull/129
[#134]: https://github.com/hyperledger-labs/go-perun/pull/134
[#148]: https://github.com/hyperledger-labs/go-perun/pull/148
[#191]: https://github.com/hyperledger-labs/go-perun/pull/191
[#196]: https://github.com/hyperledger-labs/go-perun/pull/196
[#207]: https://github.com/hyperledger-labs/go-perun/pull/207
[#218]: https://github.com/hyperledger-labs/go-perun/pull/218
[#240]: https://github.com/hyperledger-labs/go-perun/pull/240

## [0.7.0] Ganymede - 2021-07-09 [:warning:]
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
- [:boom:] **Asset holder validation** ([#111](https://github.com/hyperledger-labs/go-perun/pull/111)): Asset holder validation does no longer include adjudicator validation.
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
- <span id="warning">:warning:</span> This is a pre-release and not intended for usage with real funds.
- <span id="breaking">:boom:</span> This is a breaking change, e.g., it changes the external API.

[:warning:]: #warning
[:boom:]: #breaking

[Unreleased]: https://github.com/hyperledger-labs/go-perun/compare/v0.11.0...HEAD
[0.14.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.9.1...v0.10.0
[0.9.1]: https://github.com/hyperledger-labs/go-perun/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/hyperledger-labs/go-perun/compare/v0.5.0...v0.5.2
[0.5.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/hyperledger-labs/go-perun/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/hyperledger-labs/go-perun/releases/v0.1.0
