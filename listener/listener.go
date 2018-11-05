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
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/booster-proj/booster/sources"
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

type Provider func(context.Context) ([]core.Source, error)

type Listener struct {
	Find Provider

	s     Storage
	state *state
}

func New(s Storage) *Listener {
	return &Listener{s: s, Find: findInterfaces}
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

func (l *Listener) Run(ctx context.Context) error {
	pool := func() error {
		_ctx, cancel := context.WithTimeout(ctx, poolTimeout)
		defer cancel()

		if err := filterErr(l.Pool(_ctx)); err != nil {
			return err
		}

		log.Debug.Printf("Listener: state after pool: %+v", l.state)

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

	// Pool first
	if err := pool(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			// Exit in case of context cancelation
			return ctx.Err()
		case <-time.After(poolInterval):
			if err := pool(); err != nil {
				return err
			}
		}
	}
}

func (l *Listener) Pool(ctx context.Context) error {
	log.Debug.Printf("Listener: pooling sources...")

	if l.Find == nil {
		return &Err{fmt.Errorf("Listener requires a Find function to retrieve some sources")}
	}
	if l.state == nil {
		l.state = &state{prev: make(map[string]core.Source)}
	}

	cur, err := l.Find(ctx)
	if err != nil {
		return err
	}
	log.Debug.Printf("Pool: found %d sources", len(cur))

	curm := make(map[string]core.Source)

	// cleanup previous state
	del := []core.Source{}
	add := []core.Source{}

	for _, v := range cur {
		curm[v.ID()] = v

		if _, ok := l.state.prev[v.ID()]; !ok {
			// `v` is in cur but not in prev. Needs to be added.
			log.Debug.Printf("Pool: add source: %s", v.ID())
			add = append(add, v)
		}
	}
	for k, v := range l.state.prev {
		if _, ok := curm[k]; !ok {
			// `v` is in prev but not in cur. Has to be deleted.
			log.Debug.Printf("Pool: del source: %s", v.ID())
			del = append(del, v)
		}
	}

	l.state.prev = curm
	l.state.del = make([]core.Source, len(del))
	copy(l.state.del, del)
	l.state.add = make([]core.Source, len(add))
	copy(l.state.add, add)

	log.Debug.Printf("Pool: new state: %+v", l.state)

	return nil
}

func findInterfaces(ctx context.Context) ([]core.Source, error) {
	ifs := getFilteredInterfaces("en")
	ss := make([]core.Source, len(ifs))
	for i, v := range ifs {
		ss[i] = v
	}

	return ss, nil
}

func getFilteredInterfaces(s string) []*sources.Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		log.Error.Printf("Unable to get interfaces: %v\n", err)
		return []*sources.Interface{}
	}

	l := make([]*sources.Interface, 0, len(ifs))

	for _, v := range ifs {
		log.Debug.Printf("Inspecting interface %+v\n", v)

		if len(v.HardwareAddr) == 0 {
			log.Debug.Printf("Empty hardware address. Skipping interface...")
			continue
		}

		if s != "" && !strings.Contains(v.Name, s) {
			log.Debug.Printf("Interface name does not satisfy name requirements: must contain \"%s\"", s)
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			// If the source does not contain an error
			log.Debug.Printf("Unable to get interface addresses: %v. Skipping interface...", err)
			continue
		}
		if len(addrs) == 0 {
			log.Debug.Printf("Empty unicast/multicast address list. Skipping interface...")
			continue
		}

		l = append(l, &sources.Interface{Interface: v})
	}

	return l
}
