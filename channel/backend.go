package channel

type Backend interface {
	// ChannelID infers the channel id of a channel from its parameters. Usually,
	// this should be a hash digest of some or all fields of the parameters.
	ChannelID(*Params) ID
}
