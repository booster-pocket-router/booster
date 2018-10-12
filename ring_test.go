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

package core_test

import (
	"testing"

	"github.com/booster-proj/core"
)

func TestNew(t *testing.T) {
	r := core.NewRing(1)
	if r.Len() != 1 {
		t.Fatal("Unexpected ring size")
	}
}

func TestGetSetSource(t *testing.T) {
	r := core.NewRing(1) // create the ring
	// ensure that the ring is empty
	if s := r.Source(); s != nil {
		t.Fatalf("Unexpected ring source: %+v", s)
	}

	// set a source
	src := &srcMock{}
	r.Set(src)

	// retrieve it
	if s := r.Source(); s != src {
		t.Fatalf("Unexpected ring source: wanted %+v, found %+v", src, s)
	}
}

func TestRing(t *testing.T) {
	r0 := core.NewRing(1)
	r1 := core.NewRing(1)

	r0.Set(&srcMock{"foo"})
	r1.Set(&srcMock{"bar"})

	r0.Do(func(s core.Source) {
		if s.ID() != "foo" {
			t.Fatalf("Unexpected src content: wanted foo, found %v", s.ID())
		}
	})
	r1.Do(func(s core.Source) {
		if s.ID() != "bar" {
			t.Fatalf("Unexpected src content: wanted bar, found %v", s.ID())
		}
	})

	r := r0.Link(r1)
	if r == nil {
		t.Fatalf("Link produced nil ring")
	}

	v0 := r.Source().ID()
	v1 := r.Next().Source().ID()

	if v0 != "foo" {
		t.Fatalf("Unexpected source id: wanted foo, found %v", v0)
	}
	if v1 != "bar" {
		t.Fatalf("Unexpected source id: wanted bar, found %v", v0)
	}

	r2 := r.Next().Unlink(1)
	if r2.Len() != 1 {
		t.Fatalf("Unexpected unlinked link lenght: %v", r2.Len())
	}
	v2 := r2.Source().ID()
	if v2 != "foo" {
		t.Fatalf("Unexpected unlinked ring source id: wanted foo, found %v", v2)
	}
}
