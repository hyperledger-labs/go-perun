// Copyright (c) 2019 Chair of Applied Cryptography, Technische Universität
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
	"perun.network/go-perun/wallet"
	wtest "perun.network/go-perun/wallet/test"

	_ "perun.network/go-perun/backend/sim" // backend init
)

func TestStateSerialization(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))

	app := test.NewRandomApp(rng)
	params := test.NewRandomParams(rng, app.Def())
	state := test.NewRandomState(rng, params)

	iotest.GenericSerializerTest(t, state)
}

		}
	}
	}
}

func TestTransactionClone(t *testing.T) {
	rng := rand.New(rand.NewSource(0xDDD))
	pkgtest.VerifyClone(t, tx)
}
