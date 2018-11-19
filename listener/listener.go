/* Copyright (C) 2018 KIM KeepInMind GmbH/srl

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package listener provides a functionalities to discover and inspect
// sources.
package listener

import (
	"context" "sync"
	"time"
	"fmt"

	"github.com/booster-proj/booster/listener/provider"
	"github.com/booster-proj/core"
	"upspin.io/log"
)

// Storage describes an entity that is able to store
// and delete sources.
type Storage interface {
	Put(...core.Source)
	Del(...core.Source)
}

// Provider describes a service that is capable of providing sources
// and checking their effective internet connection using a defined
// level of confidence.
type Provider interface {
	Provide(context.Context) ([]core.Source, error)
	Check(context.Context, core.Source, provider.Confidence) error
}

type hookErr struct {
	ref     string
	network string
	address string
	err     error
}

type inspection struct {
	source core.Source
	active bool
	confidence provider.Confidence
	createdAt time.Time
}

type Listener struct {
	Provider
	s     Storage

	hooked struct {
		sync.Mutex
		ignore bool
		errors []hookErr
	}
	state struct {
		inspected map[string]inspection
	}
}

// New creates a new Listener with the provided storage, using
// as Provider the provider.Merged implementation.
func New(s Storage) *Listener {
	l := &Listener{s: s}
	l.Provider = &provider.Merged{
		ErrHook: func(ref, network, address string, err error) {
			l.hooked.Lock()
			defer l.hooked.Unlock()
			if l.hooked.ignore {
				return
			}

			log.Debug.Printf("Listener: ErrHook called from %s (net: %s, addr: %s): %v", ref, network, address, err)

			if l.hooked.errors == nil {
				l.hooked.errors = []hookErr{}
			}
			l.hooked.errors = append(l.hooked.errors, hookErr{
				ref:     ref,
				network: network,
				address: address,
				err:     err,
			})
		},
	}

	return l
}

var PollInterval = time.Second * 3
var PollTimeout = time.Second * 5

// Err is a Listener's critical error.
type Err struct {
	e error
}

func (e *Err) Error() string {
	return "critical: " + e.e.Error()
}

// filterErr either logs the error or it returns
// it, if it's critical.
func filterErr(err error) error {
	if _err, ok := err.(*Err); ok {
		return _err
	}
	if err != nil {
		log.Error.Printf("Listener error: %v", err)
	}

	return nil
}

func (l *Listener) ignoreHooks(ok bool) {
	l.hooked.Lock()
	defer l.hooked.Unlock()
	l.hooked.ignore = ok
}

// Run is a blocking function which keeps on calling Poll and waiting
// PollInterval amount of time. This function will stop with an error
// only in case of a context cancelation and in case that the Poll
// function returns with a critical error.
func (l *Listener) Run(ctx context.Context) error {
	for {
		_ctx, cancel := context.WithTimeout(ctx, PollTimeout)
		defer cancel()

		l.ignoreHooks(true)
		if err := filterErr(l.Poll(_ctx)); err != nil {
			return err
		}
		l.ignoreHooks(false)

		select {
		case <-ctx.Done():
			// Exit in case of context cancelation.
			return ctx.Err()
		case <-time.After(PollInterval):
			// Wait before polling again.
		}
	}
	return nil
}

func (l *Listener) makeInspection(src core.Source) inspection {
	err := l.Check(context.Background(), src, provider.High)
	return inspection{
		source: src,
		active: err == nil,
		confidence: provider.High,
		createdAt: time.Now(),
	}
}

// Pool queries the provider for a list of sources. It then inspect each
// new source, saving into the storage the sources that provide an active 
// internet connection and removing the ones that are no longer available.
func (l *Listener) Poll(ctx context.Context) error {
	if l.Provider == nil {
		return &Err{e: fmt.Errorf("Listener is unable to poll: no provider found")}
	}

	cur, err := l.Provide(ctx)
	if err != nil {
		return &Err{e: err}
	}


	if l.state.inspected == nil {
		l.state.inspected = make(map[string]inspection)
	}

	add := []core.Source{}
	del := []core.Source{}
	curm := make(map[string]core.Source, len(cur))

	for _, v := range cur {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		curm[v.ID()] = v

		inspection, ok := l.state.inspected[v.ID()]
		if !ok {
			// The source found is new to the listener's eyes.
			i := l.makeInspection(v)
			l.state.inspected[v.ID()] = i
			if i.active {
				add = append(add, v)
			}
			continue
		}

		// We already know this source.

		if !inspection.active {
			// The last inspection says that this source does not
			// actually provide an internet connection. Skip checking
			// for hook errors.
			continue
		}

		l.hooked.Lock()
		for _, herr := range l.hooked.errors {
			// Check if the source appears in the list of hooked
			// errors, i.e. it requires further investigation.
			if herr.ref == v.ID() {
				i := l.makeInspection(v)
				l.state.inspected[v.ID()] = i
				if !i.active {
					del = append(del, v)
				}
			}
		}
		l.hooked.Unlock()
	}

	for k, v := range l.state.inspected {
		if _, ok := curm[k]; !ok {
			// The source found was in the inspected list of items
			// but it is no longer present in the list of current
			// items. Has to be deleted.
			del = append(del, v.source)
		}
	}

	if len(add) > 0 {
		log.Info.Printf("Local provider: Adding sources: %v", add)
		l.s.Put(add...)
	}
	if len(del) > 0 {
		log.Info.Printf("Local provider: Deleting sources: %v", del)
		l.s.Del(del...)
	}

	for _, v := range del {
		// Cleanup inspections.
		delete(l.state.inspected, v.ID())
	}

	return nil
}

