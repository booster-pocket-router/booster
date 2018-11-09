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

package listener

import (
	"context"
	"sync"
	"time"

	"github.com/booster-proj/booster/listener/provider"
	"github.com/booster-proj/core"
	"upspin.io/log"
)

type Storage interface {
	Put(...core.Source)
	Del(...core.Source)
}

type state struct {
	prev map[string]core.Source
	add  []core.Source
	del  []core.Source

	updatedAt time.Time
}

type Provider interface {
	Provide(context.Context, provider.Confidence) ([]core.Source, error)
}

type hookErr struct {
	ref     string
	network string
	address string
	err     error
}

type Listener struct {
	Provider

	s     Storage
	state *state

	hooked struct {
		sync.Mutex
		ignore bool
		errors []hookErr
	}
}

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

var poolInterval = time.Second * 3
var poolTimeout = time.Second * 2

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

func (l *Listener) Run(ctx context.Context) error {
	poll := func() error {
		_ctx, cancel := context.WithTimeout(ctx, poolTimeout)
		defer cancel()

		l.ignoreHooks(true)
		if err := filterErr(l.Poll(_ctx)); err != nil {
			return err
		}
		l.ignoreHooks(false)

		log.Debug.Printf("Listener: state after poll: %+v", l.state)

		if len(l.state.del) > 0 {
			log.Info.Printf("Listener: deleting %v", l.state.del)
			l.s.Del(l.state.del...)
		}
		if len(l.state.add) > 0 {
			log.Info.Printf("Listener: adding %v", l.state.add)
			l.s.Put(l.state.add...)
		}

		return nil
	}

	// Poll first
	if err := poll(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			// Exit in case of context cancelation
			return ctx.Err()
		case <-time.After(poolInterval):
			if err := poll(); err != nil {
				return err
			}
		}
	}
}

func (l *Listener) Poll(ctx context.Context) error {
	log.Debug.Printf("Listener: pooling sources...")

	if l.state == nil {
		l.state = &state{prev: make(map[string]core.Source)}
	}

	// Find required level of confidence.
	level := provider.Low // default
	l.hooked.Lock()
	if len(l.hooked.errors) > 0 {
		// It means that at least one source is not working as
		// expected. Increase the level of confidence required
		// to find which sources are actually available.
		level = provider.High

		l.hooked.errors = []hookErr{} // reset
	}
	l.hooked.Unlock()

	log.Debug.Printf("Querying provider using confidence level: %d", level)
	cur, err := l.Provide(ctx, level)
	if err != nil {
		return err
	}
	log.Debug.Printf("Poll: found %d sources", len(cur))

	curm := make(map[string]core.Source)

	// cleanup previous state
	del := []core.Source{}
	add := []core.Source{}

	for _, v := range cur {
		curm[v.ID()] = v

		if _, ok := l.state.prev[v.ID()]; !ok {
			// `v` is in cur but not in prev. Needs to be added.
			log.Debug.Printf("Poll: add source: %s", v.ID())
			add = append(add, v)
		}
	}
	for k, v := range l.state.prev {
		if _, ok := curm[k]; !ok {
			// `v` is in prev but not in cur. Has to be deleted.
			log.Debug.Printf("Poll: del source: %s", v.ID())
			del = append(del, v)
		}
	}

	l.state.prev = curm
	l.state.del = make([]core.Source, len(del))
	copy(l.state.del, del)
	l.state.add = make([]core.Source, len(add))
	copy(l.state.add, add)

	log.Debug.Printf("Poll: new state: %+v", l.state)

	return nil
}
