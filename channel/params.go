// Copyright 2025 - See NOTICE file for copyright holders.
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
	stdio "io"
	"math/big"

	"github.com/pkg/errors"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/perunio"
)

// IDLen the length of a channelID.
const IDLen = 32

// ID represents a channelID.
type ID = [IDLen]byte

// MaxNonceLen is the maximum byte count of a nonce.
const MaxNonceLen = 32

// MinNumParts is the minimal number of participants of a channel.
const MinNumParts = 2

// Nonce is the channel parameters' nonce type.
type Nonce = *big.Int

// AuxMaxLen is the maximum byte count of the auxiliary data.
const AuxMaxLen = 256

// Aux is the channel parameters' auxiliary data type.
type Aux = [AuxMaxLen]byte

// ConvertIDToBytes converts an ID to a []byte slice.
func ConvertIDToBytes(id ID) []byte {
	return id[:]
}

// NonceFromBytes creates a nonce from a byte slice.
func NonceFromBytes(b []byte) Nonce {
	if len(b) > MaxNonceLen {
		log.Panicf("NonceFromBytes: longer than MaxNonceLen (%d/%d)", len(b), MaxNonceLen)
	}
	return new(big.Int).SetBytes(b)
}

// Zero is the default channelID.
var Zero = ID{}

var ZeroAux = Aux{}

var _ perunio.Serializer = (*Params)(nil)

// Params are a channel's immutable parameters. A channel's id is the hash of
// (some of) its parameter, as determined by the backend. All fields should be
// treated as constant.
// It should only be created through NewParams().
type Params struct {
	// ChannelID is the channel ID as calculated by the backend
	id ID
	// ChallengeDuration in seconds during disputes
	ChallengeDuration uint64
	// Parts are the channel participants
	Parts []map[wallet.BackendID]wallet.Address
	// App identifies the application that this channel is running. It is
	// optional, and if nil, signifies that a channel is a payment channel.
	App App `cloneable:"shallow"`
	// Nonce is a random value that makes the channel's ID unique.
	Nonce Nonce
	// LedgerChannel specifies whether this is a ledger channel.
	LedgerChannel bool
	// VirtualChannel specifies whether this is a virtual channel.
	VirtualChannel bool
	// Aux is an optional field that can be used to store additional information.
	Aux Aux
}

// NewParams creates Params from the given data and performs sanity checks. The
// appDef optional: if it is nil, it describes a payment channel. The channel id
// is also calculated here and persisted because it probably is an expensive
// hash operation.
func NewParams(challengeDuration uint64, parts []map[wallet.BackendID]wallet.Address, app App, nonce Nonce, ledger bool, virtual bool, aux Aux) (*Params, error) {
	if err := ValidateParameters(challengeDuration, len(parts), app, nonce); err != nil {
		return nil, errors.WithMessage(err, "invalid parameter for NewParams")
	}
	for _, ps := range parts {
		for id, p := range ps {
			if backend[p.BackendID()] == nil {
				return nil, errors.Errorf("no backend with id %d", p.BackendID())
			}
			if id != p.BackendID() {
				return nil, errors.Errorf("participant %v has wrong backend id %d", p, p.BackendID())
			}
		}
	}
	return NewParamsUnsafe(challengeDuration, parts, app, nonce, ledger, virtual, aux), nil
}

// NewParamsUnsafe creates Params from the given data and does NOT perform
// sanity checks. The channel id is also calculated here and persisted because
// it probably is an expensive hash operation.
func NewParamsUnsafe(challengeDuration uint64, parts []map[wallet.BackendID]wallet.Address, app App, nonce Nonce, ledger bool, virtual bool, aux Aux) *Params {
	p := &Params{
		ChallengeDuration: challengeDuration,
		Parts:             parts,
		App:               app,
		Nonce:             nonce,
		LedgerChannel:     ledger,
		VirtualChannel:    virtual,
		Aux:               aux,
	}

	// probably an expensive hash operation, do it only once during creation.
	id, err := CalcID(p)
	if err != nil || id == Zero {
		log.Panicf("Could not calculate channel id: %v", err)
	}
	p.id = id
	return p
}

// ID returns the channelID of this channel.
func (p *Params) ID() ID {
	return p.id
}

// ValidateProposalParameters validates all parameters that are part of the
// proposal message in the MPCPP. Checks the following conditions:
// * non-zero ChallengeDuration
// * at least two and at most MaxNumParts parts
// * appDef belongs to either a StateApp or ActionApp.
func ValidateProposalParameters(challengeDuration uint64, numParts int, app App) error {
	switch {
	case challengeDuration == 0:
		return errors.New("challengeDuration must be != 0")
	case numParts < MinNumParts:
		return errors.New("need at least two participants")
	case numParts > MaxNumParts:
		return errors.Errorf("too many participants, got: %d max: %d", numParts, MaxNumParts)
	case app == nil:
		return errors.New("app must not be nil")
	case !IsStateApp(app) && !IsActionApp(app):
		return errors.New("app must be either an Action- or StateApp")
	}
	return nil
}

// ValidateParameters checks that the arguments form valid Params. Checks
// everything from ValidateProposalParameters, and that the nonce is not nil.
func ValidateParameters(challengeDuration uint64, numParts int, app App, nonce Nonce) error {
	if err := ValidateProposalParameters(challengeDuration, numParts, app); err != nil {
		return err
	}
	if nonce == nil {
		return errors.New("nonce must not be nil")
	}
	if len(nonce.Bytes()) > MaxNonceLen {
		return errors.Errorf("nonce too long (%d > %d)", len(nonce.Bytes()), MaxNonceLen)
	}
	return nil
}

// CloneAddresses returns a clone of an Address using its binary marshaling
// implementation. It panics if an error occurs during binary (un)marshaling.
func CloneAddresses(as []map[wallet.BackendID]wallet.Address) []map[wallet.BackendID]wallet.Address {
	cloneMap := make([]map[wallet.BackendID]wallet.Address, len(as))
	for i, a := range as {
		cloneMap[i] = wallet.CloneAddressesMap(a)
	}
	return cloneMap
}

// Clone returns a deep copy of Params.
func (p *Params) Clone() *Params {
	return &Params{
		id:                p.ID(),
		ChallengeDuration: p.ChallengeDuration,
		Parts:             CloneAddresses(p.Parts),
		App:               p.App,
		Nonce:             new(big.Int).Set(p.Nonce),
		LedgerChannel:     p.LedgerChannel,
		VirtualChannel:    p.VirtualChannel,
		Aux:               p.Aux,
	}
}

// Encode uses the pkg/io module to serialize a params instance.
func (p *Params) Encode(w stdio.Writer) error {
	return perunio.Encode(w,
		p.ChallengeDuration,
		wallet.AddressMapArray{Addr: p.Parts},
		OptAppEnc{App: p.App},
		p.Nonce,
		p.LedgerChannel,
		p.VirtualChannel,
		p.Aux,
	)
}

// Decode uses the pkg/io module to deserialize a params instance.
func (p *Params) Decode(r stdio.Reader) error {
	var (
		challengeDuration uint64
		parts             wallet.AddressMapArray
		app               App
		nonce             Nonce
		ledger            bool
		virtual           bool
		aux               Aux
	)

	err := perunio.Decode(r,
		&challengeDuration,
		&parts,
		OptAppDec{App: &app},
		&nonce,
		&ledger,
		&virtual,
		&aux,
	)
	if err != nil {
		return err
	}

	_p, err := NewParams(challengeDuration, parts.Addr, app, nonce, ledger, virtual, aux)
	if err != nil {
		return err
	}
	*p = *_p

	return nil
}
