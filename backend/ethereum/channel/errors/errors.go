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

package errors

import (
	"strings"

	"github.com/pkg/errors"

	"perun.network/go-perun/client"
)

// IsChainNotReachableError checks the geth specific error that is returned in
// case the connection broke. May not work on all patforms.
func IsChainNotReachableError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "connection refused") {
		return true
	}
	return false
}

// CheckIsChainNotReachableError checks if error is a `ChainNotReachableError`
// and wraps it correctly. Returns `WithStack(err)` otherwise.
func CheckIsChainNotReachableError(err error) error {
	if IsChainNotReachableError(err) {
		return client.NewChainNotReachableError(err)
	}
	return errors.WithStack(err)
}
