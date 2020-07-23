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

package test

import (
	"sync"

	"github.com/pkg/errors"
)

type dialerList struct {
	mutex   sync.Mutex
	entries []*Dialer
}

func (l *dialerList) insert(dialer *Dialer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.entries = append(l.entries, dialer)
}

func (l *dialerList) erase(dialer *Dialer) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for i, d := range l.entries {
		if d == dialer {
			l.entries[i] = l.entries[len(l.entries)-1]
			l.entries = l.entries[:len(l.entries)-1]
			return nil
		}
	}

	return errors.New("dialer does not exist")
}

func (l *dialerList) clear() []*Dialer {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	ret := l.entries
	l.entries = nil
	return ret
}
