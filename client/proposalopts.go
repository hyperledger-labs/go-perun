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
	"perun.network/go-perun/wallet"
)

// ProposalOpts contains optional configuration instructions for channel
// proposals and channel proposal acceptance messages. Per default, no app is
// set, and a random nonce share is generated.
type ProposalOpts map[string]interface{}

var optNames = struct{ nonce, appDef, appData string }{nonce: "nonce", appDef: "appDef", appData: "appData"}

// appDef returns the option's configured app definition, or nil.
func (o ProposalOpts) appDef() wallet.Address {
	if v := o[optNames.appDef]; v != nil {
		return v.(wallet.Address)
	}
	return nil
}

// appData returns the option's configured app data, or nil.
func (o ProposalOpts) appData() channel.Data {
	if v := o[optNames.appData]; v != nil {
		return v.(channel.Data)
	}
	return nil
}

// nonce returns the option's configured nonce share, or a random nonce share.
func (o ProposalOpts) nonce() NonceShare {
	n, ok := o[optNames.nonce]
	if !ok {
		n = WithRandomNonce().nonce()
		o[optNames.nonce] = n
	}
	return n.(NonceShare)
}

// isNonce returns whether a ProposalOpts contains a manually set nonce.
func (o ProposalOpts) isNonce() bool {
	if o == nil {
		return false
	}
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

// WithNonce configures a fixed nonce share.
func WithNonce(share NonceShare) ProposalOpts {
	return ProposalOpts{optNames.nonce: share}
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

// WithApp configures an app definition and initial data.
func WithApp(appDef wallet.Address, initData channel.Data) ProposalOpts {
	return ProposalOpts{optNames.appDef: appDef, optNames.appData: initData}
}
