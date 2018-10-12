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
	"container/ring"
)

// Ring is a proxy struct around a container/ring.
// It forces to use Source as Value instead of bare interface{}.
type Ring struct {
	*ring.Ring
}

// NewRing creates a new ring with size n.
func NewRing(n int) *Ring {
	return &Ring{ring.New(n)}
}

// Source retrives the Value of the ring at this position.
func (r *Ring) Source() Source {
	if v, ok := r.Ring.Value.(Source); ok {
		return v
	}
	return nil
}

// Set sets the Value of the ring at this position to s.
func (r *Ring) Set(s Source) {
	r.Ring.Value = s
}

// Do executes f on each value of the ring.
func (r *Ring) Do(f func(Source)) {
	r.Ring.Do(func(i interface{}) {
		f(i.(Source))
	})
}

func (r *Ring) Link(s *Ring) *Ring {
	return &Ring{r.Ring.Link(s.Ring)}
}

func (r *Ring) Unlink(n int) *Ring {
	return &Ring{r.Ring.Unlink(n)}
}

func (r *Ring) Next() *Ring {
	return &Ring{r.Ring.Next()}
}

func (r *Ring) Prev() *Ring {
	return &Ring{r.Ring.Prev()}
}
