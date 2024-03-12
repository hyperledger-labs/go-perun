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

package wire

import "context"

type (
	// A Publisher allows to publish a message in a messaging network.
	Publisher interface {
		// Publish should return nil when the message was delivered (outgoing) or is
		// guaranteed to be eventually delivered (cached), depending on the goal of the
		// implementation.
		Publish(ctx context.Context, env *Envelope) error
	}

	// A Subscriber allows to subscribe Consumers, which will receive messages
	// that match a predicate.
	Subscriber interface {
		// Subscribe adds a Consumer to the subscriptions.
		// If the Consumer is already subscribed, Subscribe panics.
		Subscribe(consumer Consumer, predicate Predicate) error
	}
)
