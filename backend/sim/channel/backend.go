// Copyright 2019 - See NOTICE file for copyright holders.
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

package channel

import (
	"bytes"
	"crypto/sha256"

	"github.com/pkg/errors"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/perunio"
)

// backend implements the utility interface defined in the channel package.
type backend struct{}

var _ channel.Backend = new(backend)

// CalcID calculates a channel's ID by hashing all fields of its parameters.
func (*backend) CalcID(p *channel.Params) (id channel.ID) {
	w := sha256.New()

	// Write Parts
	for _, addr := range p.Parts {
		if err := perunio.Encode(w, addr); err != nil {
			log.Panic("Could not write to sha256 hasher")
		}
	}

	err := perunio.Encode(w, p.Nonce, p.ChallengeDuration, channel.OptAppEnc{App: p.App}, p.LedgerChannel, p.VirtualChannel)
	if err != nil {
		log.Panic("Could not write to sha256 hasher")
	}

	if copy(id[:], w.Sum(nil)) != channel.IDLen {
		log.Panic("Could not copy id")
	}
	return
}

// Sign signs `state`.
func (b *backend) Sign(addr wallet.Account, state *channel.State) (wallet.Sig, error) {
	log.WithFields(log.Fields{"channel": state.ID, "version": state.Version}).Tracef("Signing state")

	buff := new(bytes.Buffer)
	if err := state.Encode(buff); err != nil {
		return nil, errors.WithMessage(err, "pack state")
	}
	return addr.SignData(buff.Bytes())
}

// Verify verifies the signature for `state`.
func (b *backend) Verify(addr wallet.Address, state *channel.State, sig wallet.Sig) (bool, error) {
	buff := new(bytes.Buffer)
	if err := state.Encode(buff); err != nil {
		return false, errors.WithMessage(err, "pack state")
	}
	return sig.Verify(buff.Bytes(), addr)
}

// NewAsset returns a variable of type Asset, which can be used
// for unmarshalling an asset from its binary representation.
func (b *backend) NewAsset() channel.Asset {
	addr := Asset{}
	return &addr
}
