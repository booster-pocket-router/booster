/*
Copyright (C) 2018 KIM KeepInMind GmbH/srl

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

// Package core provides components useful for containing and retrieving
// entities, i.e. Sources, using a defined strategy.
package core

import (
	"context"
	"errors"
	"net"
	"sync"
)

// Dialer is a wrapper around the DialContext function.
type Dialer interface {
	// DialContext dials connections with address using the specified network.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Source represents an entity that is able to provide network connections and
// keep a set of metrics regarding the operations that is performing, or has
// performed.
type Source interface {
	// ID uniquely identifies a source.
	ID() string

	// Metrics provide information about the past usage of the source.
	Metrics() map[string]interface{}

}

// Strategy chooses a source from a ring of sources.
type Strategy func(ctx context.Context, r *Ring) (Source, error)

// RoundRobin is a naive strategy that iterates and returns each element contained in r.
func RoundRobin(ctx context.Context, r *Ring) (Source, error) {
	defer r.Next()
	return r.Source(), nil
}

// Balancer distributes work to set of sources, using a particular strategy.
// The zero value of the Balancer is ready to use and safe to be used by multiple
// gorountines.
type Balancer struct {
	mux sync.Mutex
	r   *Ring

	Strategy
}

// Get returns a Source from the balancer's source list using the predefined Strategy.
// If no Strategy was provided, Get returns a Source using RoundRobin.
func (b *Balancer) Get(ctx context.Context, blacklist ...Source) (Source, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.r == nil {
		return nil, errors.New("Empty source ring. Use Put to provide at least one source to the balancer")
	}
	if b.Strategy == nil {
		b.Strategy = RoundRobin
	}
	if len(blacklist) == 0 {
		return b.Strategy(ctx, b.r)
	}

	bl := make(map[string]interface{})
	for _, v := range blacklist {
		bl[v.ID()] = nil
	}

	var s Source
	var err error
	for i := 0; i < b.r.Len(); i++ {
		s, err = b.Strategy(ctx, b.r)
		if err != nil {
			// Avoid retring if the strategy returns an error.
			return nil, err
		}

		// Check if the source is contained in the blacklist.
		if _, ok := bl[s.ID()]; ok {
			// Skip this source
			continue
		}
		break
	}

	return s, nil
}

// Put adds ss as sources to the current balancer ring. If ss.len() == 0, Put silently returns,
// otherwise it constracts a Ring with the provided sources.
// If the balancer has already a ring, pointing lets say to 0, it adds the ring at position -1,
// preserving the balancer's ring position.
// If the balancer does not have a ring yet, the new ring will be used, making it point to
// the first source provided in the list.
func (b *Balancer) Put(ss ...Source) {
	n := len(ss)
	if n == 0 {
		return
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	// Create a new ring containing the new sources.
	s := NewRingSources(ss...)

	if b.r == nil {
		// Initialize the ring if it's still empty
		b.r = s
		return
	}

	// Make one step back. We want to keep the current position of the ring,
	// adding the sources as "tail".
	r := b.r.Prev()

	// Link the new ring to the old one. The resulting ring will point to the same
	// position as the original one.
	b.r = r.Link(s)
}

// Del removes ss from the list of sources stored by the balancer.
func (b *Balancer) Del(ss ...Source) {
	// Create a map of sources that have to be deleted (lookup O(1))
	m := make(map[string]Source)
	for _, v := range ss {
		m[v.ID()] = v
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	l := make([]Source, 0, b.r.Len())
	b.r.Do(func(s Source) {
		// Check if the identifier of this stored source is contained in the map
		// of sources that have to be removed.
		if _, ok := m[s.ID()]; !ok {
			// If this source is not contained in the map, add it to the
			// list of accepted sources.
			l = append(l, s)
		}
	})

	s := NewRingSources(l...)
	b.r = s
}

// Do executes f on each source stored in the balancer.
func (b *Balancer) Do(f func(Source)) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.r == nil {
		return
	}
	b.r.Do(f)
}


// Len reports the size of the set of sources stored in the balancer.
func (b *Balancer) Len() int {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.r == nil {
		return 0
	}

	return b.r.Len()
}
