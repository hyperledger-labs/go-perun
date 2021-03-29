// Copyright 2021 - See NOTICE file for copyright holders.
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

type (
	// OnChainTxType defines the type of on-chain transaction function names that
	// can be returned in TxTimeoutError.
	OnChainTxType int
)

// Enumeration of valid transaction types.
//
// Fund funds the channel for a given user.
// Register registers a state of the channel. The state be concluded after the challenge duration has passed.
// Progress progresses the state of the channel directly on the blockchain.
// Conclude concludes the state of a channel after it had been registered and the challenge duration has passed.
// ConcludeFinal directly concludes the finalized state of the channel without registering it.
// Withdraw withdraws the funds for a given user after the channel was concluded.
const (
	Fund OnChainTxType = iota
	Register
	Progress
	Conclude
	ConcludeFinal
	Withdraw
)

var onChainTxTypeNames = [...]string{
	"fund",
	"register",
	"progress",
	"conclude",
	"concludeFinal",
	"withdraw",
}

func (t OnChainTxType) String() string {
	return onChainTxTypeNames[t]
}
