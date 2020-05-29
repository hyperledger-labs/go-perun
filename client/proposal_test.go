// Copyright (c) 2020 Chair of Applied Cryptography, Technische Universität
// Darmstadt, Germany. All rights reserved. This file is part of go-perun. Use
// of this source code is governed by the Apache 2.0 license that can be found
// in the LICENSE file.

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProposalResponder_Accept_Nil(t *testing.T) {
	p := new(ProposalResponder)
	_, err := p.Accept(nil, *new(ProposalAcc))
	assert.Error(t, err, "context")
}
