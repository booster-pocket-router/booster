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

package core

import (
	"errors"
	"context"
	"net"
	"sync"
)

type Metrics map[string]interface{}

// Source represents an entity that is able to provide network connections and
// keep a set of metrics regarding the operations that is performing, or has
// performed.
type Source interface {
	ID() string
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	Metrics() Metrics
}

// Strategy chooses a source from a ring of sources.
type Strategy func(r *Ring) (Source, error)

// RoundRobin is a naive strategy that iterates and returns each element contained in r.
func RoundRobin(r *Ring) (Source, error) {
	defer r.Next()
	return r.Source(), nil
}

// Balancer distributes work to set of sources, using a particular strategy.
// The zero value of the Balancer is ready to use.
type Balancer struct {
	mux sync.Mutex
	r   *Ring

	Strategy
}

var errEmptyRing = errors.New("empty source ring. Use Put to provide at least one source to the balancer")

// Get returns a Source from the balancer's source list using the predefined Strategy.
// If no Strategy was provided, Get returns a Source using RoundRobin.
func (b *Balancer) Get() (Source, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.r == nil {
		return nil, errEmptyRing
	}
	if b.Strategy == nil {
		b.Strategy = RoundRobin
	}

	return b.Strategy(b.r)
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
	s := NewRing(n)
	for _, v := range ss {
		s.Set(v)
		s.Next()
	}

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
