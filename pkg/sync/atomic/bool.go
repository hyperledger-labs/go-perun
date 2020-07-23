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

// Package atomic contains extensions of "sync/atomic"
package atomic // import "perun.network/go-perun/pkg/sync/atomic"

import "sync/atomic"

// Bool is an atomically accessible boolean. Its initial state is false.
type Bool int32

// IsSet returns whether the bool is set.
func (b *Bool) IsSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }

// Set atomically sets the bool to true.
func (b *Bool) Set() { atomic.StoreInt32((*int32)(b), 1) }

// TrySet atomically sets the bool to true and returns whether it was false
// before.
func (b *Bool) TrySet() bool { return atomic.SwapInt32((*int32)(b), 1) == 0 }

// Unset atomically sets the bool to false.
func (b *Bool) Unset() { atomic.StoreInt32((*int32)(b), 0) }

// TryUnset atomically sets the bool to false and returns whether it was true
// before.
func (b *Bool) TryUnset() bool { return atomic.SwapInt32((*int32)(b), 0) == 1 }
