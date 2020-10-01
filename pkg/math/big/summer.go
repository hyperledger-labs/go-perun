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

package big

import (
	"errors"
	"math/big"
)

type (
	// A Summer can be summed up with `Sum`.
	Summer interface {
		Sum() []*big.Int
	}

	// Sum is a trivial `Summer` implementation.
	Sum []*big.Int
)

// Sum returns the receiver casted into []*big.Int.
func (s Sum) Sum() []*big.Int {
	return s
}

// AddSums sums up multiple `Summer`s and returns the `Sum`.
// Errors iff one of the `Summer`s has a different length.
func AddSums(ss ...Summer) (Sum, error) {
	if len(ss) == 0 {
		return nil, nil
	}
	sum0 := ss[0].Sum()
	sum := make(Sum, len(sum0))
	// Clone sum0 into sum.
	for j := range sum {
		sum[j] = new(big.Int).Set(sum0[j])
	}

	for i := 1; i < len(ss); i++ {
		s := ss[i].Sum()

		if len(s) != len(sum) {
			return nil, errors.New("dimension mismatch")
		}
		for j := range s {
			sum[j].Add(sum[j], s[j])
		}
	}
	return sum, nil
}

// EqualSum returns true iff two `Summer` return the same value.
func EqualSum(b0, b1 Summer) (bool, error) {
	s0, s1 := b0.Sum(), b1.Sum()
	n := len(s0)
	if n != len(s1) {
		return false, errors.New("dimension mismatch")
	}

	for i := 0; i < n; i++ {
		if s0[i].Cmp(s1[i]) != 0 {
			return false, nil
		}
	}
	return true, nil
}
