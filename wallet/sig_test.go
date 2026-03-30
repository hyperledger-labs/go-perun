// Copyright 2026 - See NOTICE file for copyright holders.
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

package wallet_test

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/wallet"
	"perun.network/go-perun/wire/perunio"
)

const (
	testDecodeBackendA wallet.BackendID = 200
	testDecodeBackendB wallet.BackendID = 201
	testDecodeBackendC wallet.BackendID = 202
	testDecodeBackendD wallet.BackendID = 203
)

var registerDecodeTestBackends sync.Once

type decodeTestAddress struct {
	id wallet.BackendID
}

func (a *decodeTestAddress) MarshalBinary() ([]byte, error) { return []byte{byte(a.id)}, nil }
func (a *decodeTestAddress) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return fmt.Errorf("unexpected address length %d", len(data))
	}
	a.id = wallet.BackendID(data[0])
	return nil
}
func (a *decodeTestAddress) String() string { return fmt.Sprintf("decode-test-%d", a.id) }
func (a *decodeTestAddress) Equal(b wallet.Address) bool {
	return a.BackendID() == b.BackendID()
}
func (a *decodeTestAddress) BackendID() wallet.BackendID { return a.id }

type decodeTestBackend struct {
	id    wallet.BackendID
	token byte
}

type replayingSparseSigReader struct {
	mask      []byte
	sig       []byte
	maskPos   int
	sigPos    int
	replaySig bool
}

func (b *decodeTestBackend) NewAddress() wallet.Address { return &decodeTestAddress{id: b.id} }
func (b *decodeTestBackend) DecodeSig(r io.Reader) (wallet.Sig, error) {
	var raw [1]byte
	if _, err := io.ReadFull(r, raw[:]); err != nil {
		return nil, err
	}
	if raw[0] != b.token {
		return nil, fmt.Errorf("unexpected signature token %q for backend %d", raw[0], b.id)
	}
	return wallet.Sig{raw[0]}, nil
}

func (*decodeTestBackend) VerifySignature([]byte, wallet.Sig, wallet.Address) (bool, error) {
	return true, nil
}

func (r *replayingSparseSigReader) Read(p []byte) (int, error) {
	if r.maskPos < len(r.mask) {
		n := copy(p, r.mask[r.maskPos:])
		r.maskPos += n
		if n < len(p) {
			m, err := r.readSig(p[n:])
			return n + m, err
		}
		return n, nil
	}
	return r.readSig(p)
}

func (r *replayingSparseSigReader) readSig(p []byte) (int, error) {
	if len(r.sig) == 0 {
		return 0, io.EOF
	}
	n := 0
	for n < len(p) {
		if r.sigPos >= len(r.sig) {
			if !r.replaySig {
				if n == 0 {
					return 0, io.EOF
				}
				return n, nil
			}
			r.sigPos = 0
		}
		m := copy(p[n:], r.sig[r.sigPos:])
		r.sigPos += m
		n += m
	}
	return n, nil
}

func ensureDecodeTestBackends() {
	registerDecodeTestBackends.Do(func() {
		wallet.SetBackend(&decodeTestBackend{id: testDecodeBackendA, token: 'a'}, int(testDecodeBackendA))
		wallet.SetBackend(&decodeTestBackend{id: testDecodeBackendB, token: 'b'}, int(testDecodeBackendB))
		wallet.SetBackend(&decodeTestBackend{id: testDecodeBackendC, token: 'x'}, int(testDecodeBackendC))
		wallet.SetBackend(&decodeTestBackend{id: testDecodeBackendD, token: 'x'}, int(testDecodeBackendD))
	})
}

func TestSigDecDecodeUsesBackendID(t *testing.T) {
	ensureDecodeTestBackends()

	var sig wallet.Sig
	backendID := testDecodeBackendB
	err := wallet.SigDec{Sig: &sig, BackendID: &backendID}.Decode(bytes.NewBuffer([]byte{'b'}))
	require.NoError(t, err)
	require.Equal(t, wallet.Sig{'b'}, sig)

	err = wallet.SigDec{Sig: &sig, BackendID: &backendID}.Decode(bytes.NewBuffer([]byte{'a'}))
	require.Error(t, err)
}

func TestDecodeSparseSigsForPartsUsesParticipantBackends(t *testing.T) {
	ensureDecodeTestBackends()

	var encoded bytes.Buffer
	require.NoError(t, perunio.Encode(&encoded, []uint8{0x03}))
	_, err := encoded.Write([]byte{'a', 'b'})
	require.NoError(t, err)

	sigs := make([]wallet.Sig, 2)
	parts := []map[wallet.BackendID]wallet.Address{
		{testDecodeBackendA: &decodeTestAddress{id: testDecodeBackendA}},
		{testDecodeBackendB: &decodeTestAddress{id: testDecodeBackendB}},
	}

	require.NoError(t, wallet.DecodeSparseSigsForParts(&encoded, &sigs, parts))
	require.Equal(t, []wallet.Sig{{'a'}, {'b'}}, sigs)
}

func TestDecodeSparseSigsForPartsFallsBackForMultiBackendParticipant(t *testing.T) {
	ensureDecodeTestBackends()

	sigs := make([]wallet.Sig, 1)
	parts := []map[wallet.BackendID]wallet.Address{{
		testDecodeBackendC: &decodeTestAddress{id: testDecodeBackendC},
		testDecodeBackendD: &decodeTestAddress{id: testDecodeBackendD},
	}}

	reader := &replayingSparseSigReader{
		mask:      []byte{0x01},
		sig:       []byte{'x'},
		replaySig: true,
	}
	require.NoError(t, wallet.DecodeSparseSigsForParts(reader, &sigs, parts))
	require.Len(t, sigs, 1)
	require.NotEmpty(t, sigs[0])
}
