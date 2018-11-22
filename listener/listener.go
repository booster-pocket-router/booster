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
	"context"
	"sync"
	"time"

	"github.com/booster-proj/booster/listener/provider"
	"github.com/booster-proj/core"
	"upspin.io/log"
)

// Storage describes an entity that is able to store
// and delete sources.
type Storage interface {
	Put(...core.Source)
	Del(...core.Source)

	Len() int
	Do(func(core.Source))
}

// Provider describes a service that is capable of providing sources
// and checking their effective internet connection using a defined
// level of confidence.
type Provider interface {
	Provide(context.Context) ([]core.Source, error)
	Check(context.Context, core.Source, provider.Confidence) error
}

// Source is a source managed by the listener
type Source struct {
	core.Source

	hooked struct {
		sync.Mutex
		Err *hookErr
	}
}

type hookErr struct {
	receivedAt time.Time
	ref        string
	network    string
	address    string
	err        error
}

type Listener struct {
	// Source provider.
	Provider

	// The location where the active sources are stored.
	s Storage
}

var PollInterval = time.Second * 3
var PollTimeout = time.Second * 5

// New creates a new Listener with the provided storage, using
// as Provider the provider.Merged implementation.
func New(s Storage) *Listener {
	return &Listener{
		s: s,
		Provider: &provider.Merged{
			ErrHook: func(ref, network, address string, err error) {
				_ = &hookErr{
					receivedAt: time.Now(),
					ref:        ref,
					network:    network,
					err:        err,
				}
				log.Debug.Printf("Listener: ErrHook called from %s (net: %s, addr: %s): %v", ref, network, address, err)
			},
		},
	}
}

// Run is a blocking function which keeps on calling Poll and waiting
// PollInterval amount of time. This function will stop with an error
// only in case of a context cancelation and in case that the Poll
// function returns with a critical error.
func (l *Listener) Run(ctx context.Context) error {
	for {
		_ctx, cancel := context.WithTimeout(ctx, PollTimeout)
		defer cancel()

		if err := l.Poll(_ctx); err != nil {
			// Just log the error
			log.Error.Println(err)
		}

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

// Diff returns respectively the list of items that has to be added and removed
// from "old" to create the same list as "cur".
func Diff(old, cur []core.Source) (add []core.Source, remove []core.Source) {
	oldm := make(map[string]core.Source, len(old))
	curm := make(map[string]core.Source, len(cur))
	for _, v := range old {
		oldm[v.ID()] = v
	}
	for _, v := range cur {
		curm[v.ID()] = v
	}

	for _, v := range old {
		// find sources to remove
		if _, ok := curm[v.ID()]; !ok {
			remove = append(remove, v)
		}
	}

	for _, v := range cur {
		// find sources to add
		if _, ok := oldm[v.ID()]; !ok {
			add = append(add, v)
		}
	}

	return
}

// Poll queries the provider for a list of sources. It then inspect each
// new source, saving into the storage the sources that provide an active
// internet connection and removing the ones that are no longer available.
func (l *Listener) Poll(ctx context.Context) error {
	// Fetch new & old data
	cur, err := l.Provide(ctx)
	if err != nil {
		return err
	}

	old := make([]core.Source, 0, l.s.Len())
	l.s.Do(func(src core.Source) {
		old = append(old, src)
	})

	// Find difference from old to cur.
	add, remove := Diff(old, cur)

	// Inspect the new ones, add them if they provide an internet connection.
	for _, v := range add {
		log.Debug.Printf("Poll: add %v?", v)
		if err := l.Check(ctx, v, provider.High); err != nil {
			log.Debug.Printf("Poll: unable to add source: %v", err)
			continue
		}
		// New source WITH active internet connection found!
		log.Info.Printf("Listener: adding (%v) to storage.", v)
		l.s.Put(v)
	}

	// Remove what has to be removed without further investigation
	for _, v := range remove {
		log.Info.Printf("Listener: removing (%v) from storage.", v)
		l.s.Del(v)
	}

	// Eventually remove the sources that contain hook errors.
	acc := make([]core.Source, 0, l.s.Len())
	l.s.Do(func(src core.Source) {
		if s, ok := src.(*Source); ok {
			s.hooked.Lock()
			if s.hooked.Err != nil {
				acc = append(acc, s.Source)
			}
			s.hooked.Unlock()
		}
	})
	for _, v := range acc {
		log.Info.Printf("Listener: removing (%v) from storage after hook error.", v)
		l.s.Del(v)
	}

	return nil
}
