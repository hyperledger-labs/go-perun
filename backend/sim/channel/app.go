package channel

import (
	"math/rand"

	"perun.network/go-perun/backend/sim/wallet"
	"perun.network/go-perun/channel"
)

type AppID struct {
	*wallet.Address
}

func (id AppID) Equal(b channel.AppID) bool {
	bTyped, ok := b.(AppID)
	if !ok {
		return false
	}

	return id.Address.Equal(bTyped.Address)
}

func (id AppID) Key() channel.AppIDKey {
	b, err := id.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return channel.AppIDKey(b)
}

func NewRandomAppID(rng *rand.Rand) AppID {
	addr := wallet.NewRandomAddress(rng)
	return AppID{addr}
}
