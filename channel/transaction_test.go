// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by a MIT-style license that can be found in
// the LICENSE file.

package channel_test

import (
	"math/rand"
	"testing"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	iotest "perun.network/go-perun/pkg/io/test"
	pkgtest "perun.network/go-perun/pkg/test"
)

func TestTransactionSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	lengths := []int{2, 5, 10, 40}
	var tests [][]bool
	for _, l := range lengths {
		tests = append(tests,
			newUniformBoolSlice(l, true), newUniformBoolSlice(l, false),
			newStripedBoolSlice(l, true), newStripedBoolSlice(l, false))
		for i := 0; i < l; i++ {
			tests = append(tests, newAlmostUniformBoolSlice(i, l, true))
			tests = append(tests, newAlmostUniformBoolSlice(i, l, false))
		}
	}

	for _, tt := range tests {
		tx := test.NewRandomTransaction(rng, tt)
		iotest.GenericSerializerTest(t, tx)
	}

	tx := new(channel.Transaction)
	iotest.GenericSerializerTest(t, tx)
}

// newUniformBoolSlice generates a slice long size with all the elements set to choice
func newUniformBoolSlice(size int, choice bool) []bool {
	uniform := make([]bool, size)
	for i := range uniform {
		uniform[i] = choice
	}
	return uniform
}

// newAlmostUniformBoolSlice creates []bool which has choice at indexChosen and all the others indexes are !choice
func newAlmostUniformBoolSlice(indexChosen int, size int, choice bool) []bool {
	almostUniform := make([]bool, size)
	for i := range almostUniform {
		if i == indexChosen {
			almostUniform[i] = choice
		} else {
			almostUniform[i] = !choice
		}
	}
	return almostUniform
}

// newStripedBoolSlice creates an array []bool of length == size in which all the even indexes are set to choice
func newStripedBoolSlice(size int, choice bool) []bool {
	striped := make([]bool, size)
	for i := range striped {
		striped[i] = choice
		choice = !choice
	}
	return striped
}

func TestTransactionClone(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDD))
	size := int(rng.Int31n(5)) + 2
	testmask := newUniformBoolSlice(size, true)
	tx := *test.NewRandomTransaction(rng, testmask)
	pkgtest.VerifyClone(t, tx)
}
