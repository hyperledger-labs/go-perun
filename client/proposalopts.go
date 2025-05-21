// Copyright 2020 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"crypto/rand"
	"io"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
)

// ProposalOpts contains optional configuration instructions for channel
// proposals and channel proposal acceptance messages. Per default, NoApp and
// NoData is set, and a random nonce share is generated.
type ProposalOpts map[string]interface{}

var optNames = struct{ nonce, app, appData, fundingAgreement, aux string }{nonce: "nonce", app: "app", appData: "appData", fundingAgreement: "fundingAgreement", aux: "aux"}

// App returns the option's configured app.
func (o ProposalOpts) App() channel.App {
	if v := o[optNames.app]; v != nil {
		app, ok := v.(channel.App)
		if !ok {
			log.Panicf("wrong type: expected channel.App, got %T", v)
		}
		return app
	}
	return channel.NoApp()
}

// AppData returns the option's configured app data.
func (o ProposalOpts) AppData() channel.Data {
	if v := o[optNames.appData]; v != nil {
		data, ok := v.(channel.Data)
		if !ok {
			log.Panicf("wrong type: expected channel.Data, got %T", v)
		}
		return data
	}
	return channel.NoData()
}

// SetsApp returns whether an app and data have been explicitly set.
func (o ProposalOpts) SetsApp() bool {
	_, ok := o[optNames.app]
	return ok
}

func (o ProposalOpts) isFundingAgreement() bool {
	_, ok := o[optNames.fundingAgreement]
	return ok
}

// fundingAgreement returns the funding-agreement that was set by
// `WithFundingAgreement` and panics otherwise.
func (o ProposalOpts) fundingAgreement() channel.Balances {
	a, ok := o[optNames.fundingAgreement]
	if !ok {
		panic("Option FundingAgreement not set")
	}
	bals, ok := a.(channel.Balances)
	if !ok {
		log.Panicf("wrong type: expected channel.Balances, got %T", a)
	}
	return bals
}

// nonce returns the option's configured nonce share, or a random nonce share.
func (o ProposalOpts) nonce() NonceShare {
	n, ok := o[optNames.nonce]
	if !ok {
		n = WithRandomNonce().nonce()
		o[optNames.nonce] = n
	}
	share, ok := n.(NonceShare)
	if !ok {
		log.Panicf("wrong type: expected NonceShare, got %T", n)
	}
	return share
}

// aux returns the option's configured auxiliary data.
func (o ProposalOpts) aux() channel.Aux {
	a, ok := o[optNames.aux]
	if !ok {
		return channel.ZeroAux
	}
	aux, ok := a.(channel.Aux)
	if !ok {
		log.Panicf("wrong type: expected []byte, got %T", a)
	}
	return aux
}

// isNonce returns whether a ProposalOpts contains a manually set nonce.
func (o ProposalOpts) isNonce() bool {
	_, ok := o[optNames.nonce]
	return ok
}

func union(opts ...ProposalOpts) ProposalOpts {
	ret := ProposalOpts{}
	for _, opt := range opts {
		for k, v := range opt {
			_, ok := ret[k]
			if ok {
				log.Panicf("ProposalOpts: duplicate %s option", k)
			}
			ret[k] = v
		}
	}
	return ret
}

// WithFundingAgreement configures a fixed funding agreement.
func WithFundingAgreement(alloc channel.Balances) ProposalOpts {
	return ProposalOpts{optNames.fundingAgreement: alloc}
}

// WithNonce configures a fixed nonce share.
func WithNonce(share NonceShare) ProposalOpts {
	return ProposalOpts{optNames.nonce: share}
}

// WithAux configures an auxiliary data field.
func WithAux(aux channel.Aux) ProposalOpts {
	return ProposalOpts{optNames.aux: aux}
}

// WithNonceFrom reads a nonce share from a reader (should be random stream).
func WithNonceFrom(r io.Reader) ProposalOpts {
	var share NonceShare
	if _, err := io.ReadFull(r, share[:]); err != nil {
		log.Panic("Failed to read nonce share", err)
	}
	return WithNonce(share)
}

// WithRandomNonce creates a nonce from crypto/rand.Reader.
func WithRandomNonce() ProposalOpts {
	return WithNonceFrom(rand.Reader)
}

// WithApp configures an app and initial data.
func WithApp(app channel.App, initData channel.Data) ProposalOpts {
	return ProposalOpts{optNames.app: app, optNames.appData: initData}
}

// WithoutApp configures a NoApp and NoData.
func WithoutApp() ProposalOpts {
	return WithApp(channel.NoApp(), channel.NoData())
}
