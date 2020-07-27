<h1 align="center"><br>
    <a href="https://perun.network/"><img src=".assets/logo.png" alt="Perun" width="196"></a>
<br></h1>

<h4 align="center">Perun Blockchain-Agnostic State Channels Framework</h4>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/perun-network/go-perun"><img src="https://goreportcard.com/badge/github.com/perun-network/go-perun" alt="Go report: A+"></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0.txt"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License: Apache 2.0"></a>
  <a href="https://travis-ci.org/perun-network/go-perun"><img src="https://travis-ci.org/perun-network/go-perun.svg?branch=dev" alt="TravisCI build status"></a>
  <a href="https://pkg.go.dev/perun.network/go-perun?status.svg"> <img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="pkg.go.dev docs"></a>
</p>

_go-perun_ is a Go implementation of the [Perun state channel protocols](https://perun.network/) ([introduction paper](https://perun.network/pdf/Perun2.0.pdf)).
The perun protocols provide payment and general state channel functionality to all existing blockchains that feature smart contracts.
As a blockchain scalability solution, payment and state channels reduce transaction costs and increase the system throughput by executing incremental transactions off-chain.
The Perun protocols have been proven cryptographically secure in the UC-framework.
They are blockchain-agnostic and only rely on a blockchain's capability to execute smart contracts.

## Security Disclaimer

_go-perun_ is still alpha software.
It should not be used in production.
The [current release](https://github.com/perun-network/go-perun/releases) is not intended to have any practical use, and should only give potential users a general impression and invite feedback.
The authors take no responsibility for any loss of digital assets or other damage caused by the use of this software.
**Do not use this software with real funds**.

## Getting Started

Running _go-perun_ requires a working Go distribution (version 1.14 or higher).
```sh
# Clone the repository into a directory of your choice
git clone https://github.com/perun-network/go-perun.git
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

More in-depth documentation can be found in the [github wiki pages](https://github.com/perun-network/go-perun/wiki)
and on [go-perun's pkg.go.dev site](https://pkg.go.dev/perun.network/go-perun).

## Features

_go-perun_ currently supports all features needed for two party payment channels.
The following features are currently provided:
* Two-party ledger state channels
* Cooperatively settling
* Ledger channel disputes
* Dispute watchtower
* Data persistence

The following features are planned for future releases:
* Generalized two-party ledger channels (sub-channels)
* Virtual two-party channels (direct dispute)
* Virtual two-party channels (indirect dispute)
* Multi-party ledger channels
* Virtual multi-party channels (direct dispute)
* Cross-blockchain virtual channels (indirect dispute)

### Backends

There are multiple **blockchain backends** available as part of the current release: Ethereum (`backend/ethereum`), and a simulated, ideal blockchain backend (`backend/sim`).
A backend is automatically initialized when its top-level package `backend/<name>` is imported.
The Ethereum smart contracts can be found in our [contracts-eth](https://github.com/perun-network/contracts-eth) repository.

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
	var w wallet.Wallet

	// create state channel network client
	c := client.New(perunID, bus, funder, adjudicator, w)

	// choose how to react to incoming channel proposals
	var proposalHandler client.ProposalHandler
	// choose how to react to incoming channel update requests
	var updateHandler client.UpdateHandler
	// start incoming request handler
	go c.Handle(proposalHandler, updateHandler)

	// propose a new channel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	ch, err := c.ProposeChannel(ctx, &client.ChannelProposal{
		// details of channel proposal, like peers, app, initial balances, challenge duration...
	})
	if err != nil { /* handle error */ }

	// start watchtower
	go func() {
		err := ch.Watch()
		log.Info("Watcher returned with error ", err)
	}()

	// send a channel update request to the other channel peer(s)
	err = ch.Update(ctx, client.ChannelUpdate{
		// details of channel update: new State, actor index, ...
	})
	if err != nil { /* handle error */ }

	// send further updates and finally, settle/close the channel
	if err := ch.Settle(ctx); err != nil { /* handle error */ }
}
```

For a full-fledged example, have a look at our CLI Demo [perun-eth-demo](https://github.com/perun-network/perun-eth-demo).
Go mobile wrappers for <img src="https://developer.android.com/images/brand/Android_Robot.svg?hl=de" width="25" alt="Android"> and iOS App development can be found at [perun-eth-mobile](https://github.com/perun-network/perun-eth-mobile).

## Funding

This project is currently being developed by a group of dedicated hackers at the Applied Cryptography research group at Technische Universität Darmstadt, Germany.
We thank the German Federal Ministry of Education and Research (BMBF) for their funding through the StartUpSecure grants program as well as the German Science Foundation (DFG), the Foundation for Polish Science (FNP) and the Ethereum Foundation for their support in the research that preceded this implementation.

## Copyright

Copyright 2020 - See [NOTICE file](NOTICE) for copyright holders.  
Use of the source code is governed by the Apache 2.0 license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
