<h1 align="center"><br>
    <a href="https://perun.network/"><img src=".assets/logo.png" alt="Perun" width="196"></a>
<br></h1>

<h4 align="center">Perun Blockchain-Agnostic State Channels Framework</h4>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/perun-network/go-perun"><img src="https://goreportcard.com/badge/github.com/perun-network/go-perun" alt="Go report: A+"></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0.txt"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License: Apache 2.0"></a>
  <a href="https://travis-ci.org/perun-network/go-perun"><img src="https://travis-ci.org/perun-network/go-perun.svg?branch=dev" alt="TravisCI build status"></a>
</p>

_go-perun_ is a Go implementation of the [Perun state channel protocols](https://perun.network/) ([introduction paper](https://perun.network/pdf/Perun2.0.pdf)).
The perun protocols provide payment and general state channel functionality to all existing blockchains that feature smart contracts.
As a blockchain scalability solution, payment and state channels reduce transaction costs and increase the system throughput by executing incremental transactions off-chain.
The Perun protocols have been proven cryptographically secure in the UC-framework.
They are blockchain-agnostic and only rely on a blockchain's capability to execute smart contracts.

## Security Disclaimer

_go-perun_ is still alpha software.
It should not be used in production.
The current release, _Belinda_, is not intended to have any practical use, and should only give potential users a general impression and invite feedback.
Some essential features, such as data persistence, are not yet implemented.
The authors take no responsibility for any loss of digital assets or other damage caused by the use of this software.
**Do not use this software with real funds**.

## Getting Started

Running _go-perun_ requires a working Go distribution (version 1.13 or higher).
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

### Backends

There are multiple backends available as part of the current release: Ethereum (`backend/ethereum`), and a simulated, ideal blockchain backend (`backend/sim`).
A backend is automatically initialized when its `wallet` and `channel` packages are imported.
The Ethereum smart contracts can be found in our [contracts-eth](https://github.com/perun-network/contracts-eth) repository.

Logging and networking capabilities can also be injected by the user.
A default [logrus](https://github.com/sirupsen/logrus) implementation of the `log.Logger` interface can be set using [`log/logrus.Set`](log/logrus/logrus.go#L44).
The Perun framework relies on `peer.Dialer` and `peer.Listener` implementations for networking.
_go-perun_ is distributed with TCP and Unix socket implementations for testing
purposes, which can be found in package `peer/net`.

## Features

_go-perun_ currently only supports a reduced set of features compared to the full protocols.
The following table shows the list of features needed for the minimal, secure, production-ready software.

| Feature                          | Ariel (v0.1.0)     | Belinda (v0.2.0)   |
| -------------------------------- | ------------------ | ------------------ |
| Two-party ledger state channels  | :heavy_check_mark: | :heavy_check_mark: |
| Cooperatively settling           | :heavy_check_mark: | :heavy_check_mark: |
| Ledger channel dispute           | :x:                | :heavy_check_mark: |
| Dispute watchtower               | :x:                | :heavy_check_mark: |
| Data persistence                 | :x:                | :x:                |

The following features are planned after the above features have been implemented:
* Generalized two-party ledger channels
* Virtual channels (direct dispute)
* Multi-party ledger channels
* Virtual two-party channels (indirect dispute)
* Virtual multi-party channels (direct dispute)
* Cross-blockchain virtual channels (indirect dispute)

## API Primer

In essence, _go-perun_ provides a state channel network client, akin to ethereum's `ethclient` package, to interact with a state channels network.
Once the client has been set up, it can be used to propose channels to other network peers, send updates on those channels and eventually settle them.
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
	// other imports
)

func main() {
	// setup networking
	var dialer peer.Dialer
	var listener peer.Listener
	// setup blockchain interaction
	var funder channel.Funder
	var adjudicator channel.Adjudicator
	// setup off-chain identity
	var identity peer.Identity
	// setup wallet for channel accounts
	var w wallet.Wallet

	// create state channel network client
	c := client.New(identity, dialer, funder, adjudicator, w)
	// optionally start listening for incoming connections
	go c.Listen(listener)

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

## Acknowledgements

This project is currently being developed by a group of dedicated hackers at the Applied Cryptography research group at Technische Universität Darmstadt, Germany.
We thank the German Federal Ministry of Education and Research (BMBF) for their funding through the StartUpSecure grants program as well as the German Science Foundation (DFG), the Foundation for Polish Science (FNP) and the Ethereum Foundation for their support in the research that preceded this implementation.

## Copyright

Copyright &copy; 2020 Chair of Applied Cryptography, Technische Universität Darmstadt, Germany.
All rights reserved.
Use of the source code is governed by the Apache 2.0 license that can be found in the [LICENSE file](LICENSE).

Contact us at [info@perun.network](mailto:info@perun.network).
