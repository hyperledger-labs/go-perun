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

package channel

import (
	"log"
	"sync"

	"github.com/pkg/errors"
)

// appRegistry is the global registry for `AppResolver`s.
var appRegistry = appReg{singles: make(map[AppIDKey]App)}

type AppIDKey string

type appReg struct {
	sync.RWMutex
	resolvers  []appRegEntry
	singles    map[AppIDKey]App
	defaultRes AppResolver
}

// AppIDPredicate is a function for filtering app identifiers.
type AppIDPredicate = func(AppID) bool

type appRegEntry struct {
	pred AppIDPredicate
	res  AppResolver
}

// Resolve is a global wrapper call to the global `appRegistry`.
// This function is intended to resolve app definitions coming in on the wire.
func Resolve(def AppID) (App, error) {
	appRegistry.RLock()
	defer appRegistry.RUnlock()
	if def == nil {
		log.Panic("resolving nil address")
	}
	if app, ok := appRegistry.singles[def.Key()]; ok {
		return app, nil
	}
	for _, e := range appRegistry.resolvers {
		if e.pred(def) {
			return e.res.Resolve(def)
		}
	}
	if appRegistry.defaultRes == nil {
		return nil, errors.Errorf("def %v could not be resolved and no default resolver set", def)
	}
	return appRegistry.defaultRes.Resolve(def)
}

// RegisterAppResolver appends the given `AppIDPredicate` and `AppResolver` to
// the global `appRegistry`.
func RegisterAppResolver(pred AppIDPredicate, appRes AppResolver) {
	appRegistry.Lock()
	defer appRegistry.Unlock()

	if pred == nil || appRes == nil {
		log.Panic("nil AppIDPredicate or AppResolver")
	}

	appRegistry.resolvers = append(appRegistry.resolvers, appRegEntry{pred, appRes})
}

// RegisterApp registers a single app for a single address.
func RegisterApp(app App) {
	appRegistry.Lock()
	defer appRegistry.Unlock()

	if app == nil || app.Def() == nil {
		log.Panic("nil Address or App")
	}

	appRegistry.singles[app.Def().Key()] = app
}

// RegisterDefaultApp allows to specify a default `AppResolver` which is used by
// the `AppRegistry` if no predicate matches. It must be set during the
// initialization of the program, before any app is resolved.
func RegisterDefaultApp(appRes AppResolver) {
	appRegistry.Lock()
	defer appRegistry.Unlock()
	if appRes == nil {
		log.Panic("nil AppResolver")
	}
	if appRegistry.defaultRes != nil {
		log.Panic("default resolver already set")
	}
	appRegistry.defaultRes = appRes
}
