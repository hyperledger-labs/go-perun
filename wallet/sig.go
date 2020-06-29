// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package wallet

import (
	"bytes"
	"io"
	"math"

	"github.com/pkg/errors"

	perunio "perun.network/go-perun/pkg/io"
)

// Sig is a single signature
type Sig = []byte

// CloneSigs returns a deep copy of a slice of signatures
func CloneSigs(sigs []Sig) []Sig {
	if sigs == nil {
		return nil
	}
	clonedSigs := make([]Sig, len(sigs))
	for i, sig := range sigs {
		if sig != nil {
			clonedSigs[i] = bytes.Repeat(sig, 1)
		}
	}
	return clonedSigs
}

var _ perunio.Decoder = SigDec{}

// SigDec is a helper type to decode signatures.
type SigDec struct {
	Sig *Sig
}

// Decode decodes a single signature.
func (s SigDec) Decode(r io.Reader) (err error) {
	*s.Sig, err = DecodeSig(r)
	return err
}

// EncodeSparseSigs encodes a collection of signatures in the form ( mask, sig, sig, sig, ...)
func EncodeSparseSigs(w io.Writer, sigs []Sig) error {
	n := len(sigs)

	// Encode mask
	mask := make([]uint8, int(math.Ceil(float64(n)/8.0)))
	for i, sig := range sigs {
		if sig != nil {
			mask[i/8] |= 0x01 << (i % 8)
		}
	}
	if err := perunio.Encode(w, mask); err != nil {
		return errors.WithMessage(err, "encoding mask")
	}

	// Encode signatures
	for _, sig := range sigs {
		if sig != nil {
			if err := perunio.Encode(w, sig); err != nil {
				return errors.WithMessage(err, "encoding signature")
			}
		}
	}
	return nil
}

// DecodeSparseSigs decodes a collection of signatures in the form (mask, sig, sig, sig, ...)
func DecodeSparseSigs(r io.Reader, sigs *[]Sig) (err error) {
	masklen := int(math.Ceil(float64(len(*sigs)) / 8.0))
	mask := make([]uint8, masklen)

	//Decode mask
	if err = perunio.Decode(r, &mask); err != nil {
		return errors.WithMessage(err, "decoding mask")
	}

	//Decoding mask's signatures
	for maskIdx, sigIdx := 0, 0; maskIdx < len(mask); maskIdx++ {
		for bitIdx := 0; bitIdx < 8 && sigIdx < len(*sigs); bitIdx, sigIdx = bitIdx+1, sigIdx+1 {
			if ((mask[maskIdx] >> bitIdx) % 2) == 0 {
				(*sigs)[sigIdx] = nil
			} else {
				(*sigs)[sigIdx], err = DecodeSig(r)
				if err != nil {
					return errors.WithMessagef(err, "decoding signature %d", sigIdx)
				}
			}
		}
	}
	return nil
}
