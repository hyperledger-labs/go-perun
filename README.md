<h1 align="center"><br>
    <a href="https://perun.network/"><img src=".assets/logo.png" alt="Perun" width="196"></a>
<br></h1>

<h4 align="center">Perun Blockchain-Agnostic State Channels Framework</h4>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/hyperledger-labs/go-perun"><img src="https://goreportcard.com/badge/github.com/hyperledger-labs/go-perun" alt="Go report: A+"></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0.txt"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License: Apache 2.0"></a>
  <a href="https://github.com/hyperledger-labs/go-perun/actions/workflows/ci.yml"><img src="https://github.com/hyperledger-labs/go-perun/actions/workflows/ci.yml/badge.svg" alt="TravisCI build status"></a>
  <a href="https://pkg.go.dev/perun.network/go-perun?status.svg"> <img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="pkg.go.dev docs"></a>
  <a href="https://labs.hyperledger.org/perun-doc/"> <img src=".assets/read_the_docs.svg" alt="Read the Docs"></a>
</p>

_go-perun_ is a Go implementation of the [Perun state channel protocols](https://perun.network/) ([introduction paper](https://perun.network/wp-content/uploads/Perun2.0.pdf)).
The perun protocols provide payment and general state channel functionality to all existing blockchains that feature smart contracts.
As a blockchain scalability solution, payment and state channels reduce transaction costs and increase the system throughput by executing incremental transactions off-chain.
The Perun protocols have been proven cryptographically secure in the UC-framework.
They are blockchain-agnostic and only rely on a blockchain's capability to execute smart contracts.

## Security Disclaimer

This software is still under development.
The authors take no responsibility for any loss of digital assets or other damage caused by the use of it.

## Getting Started

Running _go-perun_ requires a working [Go distribution](https://golang.org), see `go.mod` for the required version.
```sh
# Clone the repository into a directory of your choice
git clone https://github.com/hyperledger-labs/go-perun.git
# Or directly download it with go
# go get -d perun.network/go-perun
cd go-perun
# Run the unit tests
go test ./...
```

You can import _go-perun_ in your project like this:
```go
import "perun.network/go-perun/client"
```

_go-perun_ implements the core state channel protocol in a blockchain-agnostic fashion by following the dependency inversion principle.
For this reason, a blockchain backend has to be chosen and blockchain-specific initializations need to be executed at program startup.

### Documentation

In-depth documentation can be found at [perun-doc](https://labs.hyperledger.org/perun-doc/)
and on [go-perun's pkg.go.dev site](https://pkg.go.dev/perun.network/go-perun).

Our [walkthrough tutorial](http://tutorial.perun.network) describes how _go-perun_ is used to build a simple micro-payment application on top of Ethereum.

A bunch of examples can be found at [perun-examples](https://github.com/perun-network/perun-examples).

## Features

_go-perun_ currently provides two-party channels with the following features:
* Generalized state channels, including app/sub-channels
* Dispute watchtower
* Data persistence
* Virtual payment channels (direct dispute)
* Multi-ledger channels

The following features are planned for future releases:
* Virtual payment channels (indirect dispute)
* Virtual state channels
* Multi-party channels
* Multi-ledger app/virtual channels

### Backends

There are multiple **blockchain backends** available. A backend is automatically initialized when its top-level package `backend/<name>` is imported.
- **Ethereum.** The Ethereum backend is available at [perun-eth-backend](https://github.com/perun-network/perun-eth-backend/).
- **Polkadot.** The Polkadot backend is available at [perun-polkadot-backend](https://github.com/perun-network/perun-polkadot-backend).
- **Cosmos.** The Cosmos backend is available at [perun-cosmwasm-backend](https://github.com/hyperledger-labs/perun-cosmwasm-backend).
- **SimBackend.** The SimBackend represents an ideal blockchain backend (`backend/sim`) implementation that can be used for testing.

**Logging and networking** capabilities can also be injected by the user.
A default [logrus](https://github.com/sirupsen/logrus) implementation of the `log.Logger` interface can be set using `log/logrus.Set`.
The Perun framework relies on a user-injected `wire.Bus` for inter-peer communication.
_go-perun_ ships with the `wire/net.Bus` implementation for TCP and Unix sockets.

**Data persistence** can be enabled to continuously persist new states and signatures.
There are currently three persistence backends provided, namely, a test backend for testing purposes, an in-memory key-value persister and a [LevelDB](https://github.com/syndtr/goleveldb) backend.

## API Primer

In essence, _go-perun_ provides a state channel network client, akin to ethereum's `ethclient` package, to interact with a state channels network.
Once the client has been set up, it can be used to propose channels to other network peers, accept channel proposals, send updates on those channels and eventually settle them.

A minimal, illustrative usage is as follows
```go
package main

import (
	"context"
	"time"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/client"
	"perun.network/go-perun/log"
	"perun.network/go-perun/peer"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/watcher"
	"perun.network/go-perun/wire"
	// other imports
)

func main() {
	// setup blockchain interaction
	var funder channel.Funder
	var adjudicator channel.Adjudicator
	// setup perun network identity
	var perunID wire.Address
	// setup communication bus
	var bus wire.Bus
	// setup wallet for channel accounts
	var wallet wallet.Wallet
	// setup channel watcher
	var watcher watcher.Watcher

	// create state channel network client
	c := client.New(perunID, bus, funder, adjudicator, wallet, watcher)

	// choose how to react to incoming channel proposals
	var proposalHandler client.ProposalHandler
	// choose how to react to incoming channel update requests
	var updateHandler client.UpdateHandler
	// start incoming request handler
	go c.Handle(proposalHandler, updateHandler)

	// propose a new channel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ch, err := c.ProposeChannel(ctx, client.NewLedgerChannelProposal(
		// details of channel proposal, like peers, app, initial balances, challenge duration...
	))
	if err != nil { /* handle error */ }

	// start watchtower
	go func() {
		err := ch.Watch()
		log.Info("Watcher returned with error ", err)
	}()

	// send a channel update request to the other channel peer(s)
	err = ch.Update(ctx, func(s *channel.State) {
		// update state s, e.g., moving funds or changing app data
	})
	if err != nil { /* handle error */ }

	// send further updates and finally, settle/close the channel
	if err := ch.Settle(ctx); err != nil { /* handle error */ }
}
```

For a full-fledged example, have a look at our CLI Demo [perun-eth-demo](https://github.com/perun-network/perun-eth-demo).
Go mobile wrappers for <img src="https://developer.android.com/images/brand/Android_Robot.svg?hl=de" width="25" alt="Android"> and iOS App development can be found at [perun-eth-mobile](https://github.com/perun-network/perun-eth-mobile).

## Funding

This project has been started at the Chair of Applied Cryptography at Technische Universität Darmstadt, Germany, and is currently developed and maintained by a group of dedicated hackers from PolyCrypt GmbH.
We thank the German Federal Ministry of Education and Research (BMBF) for their funding through the StartUpSecure grants program as well as the German Science Foundation (DFG), the Foundation for Polish Science (FNP) and the Ethereum Foundation for their support in the research that preceded this implementation.

## Copyright

Copyright 2022 - See [NOTICE file](NOTICE) for copyright holders.
Use of the source code is governed by the Apache 2.0 license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
