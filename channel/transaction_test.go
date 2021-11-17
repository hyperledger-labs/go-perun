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

package channel_test

import (
	"testing"

	"perun.network/go-perun/channel"
	"perun.network/go-perun/channel/test"
	iotest "polycry.pt/poly-go/io/test"
	pkgtest "polycry.pt/poly-go/test"
)

func TestTransactionSerialization(t *testing.T) {
	rng := pkgtest.Prng(t)
	lengths := []int{2, 5, 10, 40}
	tests := [][]bool{}
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

// newUniformBoolSlice generates a slice long size with all the elements set to choice.
func newUniformBoolSlice(size int, choice bool) []bool {
	uniform := make([]bool, size)
	for i := range uniform {
		uniform[i] = choice
	}
	return uniform
}

// newAlmostUniformBoolSlice creates []bool which has choice at indexChosen and all the others indexes are !choice.
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

// newStripedBoolSlice creates an array []bool of length == size in which all the even indexes are set to choice.
func newStripedBoolSlice(size int, choice bool) []bool {
	striped := make([]bool, size)
	for i := range striped {
		striped[i] = choice
		choice = !choice
	}
	return striped
}

func TestTransactionClone(t *testing.T) {
	rng := pkgtest.Prng(t)
	size := int(rng.Int31n(5)) + 2
	testmask := newUniformBoolSlice(size, true)
	tx := *test.NewRandomTransaction(rng, testmask)
	pkgtest.VerifyClone(t, tx)
}
