// Copyright 2022 - See NOTICE file for copyright holders.
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
	"encoding/json"
	"fmt"

	"perun.network/go-perun/wallet"
)

type (
	jsonParams struct {
		ChallengeDuration uint64
		Parts             []json.RawMessage
		Nonce             Nonce
		LedgerChannel     bool
		VirtualChannel    bool
	}

	jsonState struct {
		ID         *ID            `json:"id"`
		Version    *uint64        `json:"version"`
		Allocation jsonAllocation `json:"allocation"`
		IsFinal    *bool          `json:"final"`
	}

	jsonAllocation struct {
		Assets   []json.RawMessage `json:"assets"`
		Balances *Balances         `json:"balances"`
		Locked   *[]SubAlloc       `json:"locked"`
	}
)

// UnmarshalJSON decodes the given json data into the Params.
func (p *Params) UnmarshalJSON(data []byte) error {
	var jp jsonParams
	if err := json.Unmarshal(data, &jp); err != nil {
		return fmt.Errorf("unmarshaling into jsonParams: %w", err)
	}

	parts := make([]wallet.Address, 0, len(jp.Parts))
	for i, rawPart := range jp.Parts {
		a := wallet.NewAddress()
		if err := json.Unmarshal(rawPart, &a); err != nil {
			return fmt.Errorf("unmarshaling participant[%d]: %w", i, err)
		}
		parts = append(parts, a)
	}

	params, err := NewParams(
		jp.ChallengeDuration, parts, NoApp(), jp.Nonce,
		jp.LedgerChannel, jp.VirtualChannel)
	if err != nil {
		return err
	}
	*p = *params
	return nil
}

// UnmarshalJSON decodes the given json data into the State.
func (s *State) UnmarshalJSON(data []byte) error {
	js := jsonState{
		ID:      &s.ID,
		Version: &s.Version,
		IsFinal: &s.IsFinal,
		Allocation: jsonAllocation{
			Balances: &s.Balances,
			Locked:   &s.Locked,
		},
	}
	if err := json.Unmarshal(data, &js); err != nil {
		return fmt.Errorf("unmarshaling into jsonState: %w", err)
	}

	s.App = NoApp()
	s.Data = NoData()
	s.Assets = make([]Asset, 0, len(js.Allocation.Assets))
	for i, rawAsset := range js.Allocation.Assets {
		a := NewAsset()
		if err := json.Unmarshal(rawAsset, &a); err != nil {
			return fmt.Errorf("unmarshaling asset[%d]: %w", i, err)
		}
		s.Assets = append(s.Assets, a)
	}
	return s.Allocation.Valid()
}
