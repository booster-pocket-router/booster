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
	"fmt"

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
	src := &mock{}
	r.Set(src)

	// retrieve it
	if s := r.Source(); s != src {
		t.Fatalf("Unexpected ring source: wanted %+v, found %+v", src, s)
	}
}

func TestNextPrev(t *testing.T) {
	r := core.NewRing(4)
	for i := 0; i < r.Len(); i++ {
		s := fmt.Sprintf("%d", i)
		r.Set(&mock{s})
		r.Next()
	}

	tt := []string { "0", "1", "2", "3" }
	i := 0
	r.Do(func(s core.Source) {
		v := tt[i]
		if v != s.ID() {
			t.Fatalf("Unexpected id value (%d): wanted %v, found %v", i, v, s.ID())
		}
		i++
	})

	// We should be at element 0 at this point.
	if r.Prev().Source().ID() != "3" {
		t.Fatalf("Unexpected Prev id: wanted 3, found %s", r.Source().ID())
	}
}

func TestLink(t *testing.T) {
	n := 2
	r0 := core.NewRing(n)
	r1 := core.NewRing(n)

	for i := 0; i < n; i++ {
		s0 := fmt.Sprintf("%d", i)
		s1 := fmt.Sprintf("%d", i + n)
		r0.Set(&mock{s0})
		r1.Set(&mock{s1})

		r0.Next()
		r1.Next()
	}

	t.Log("r0 before link:")
	r0.Do(func(s core.Source) {
		t.Log(s)
	})
	t.Log("r1 before link:")
	r1.Do(func(s core.Source) {
		t.Log(s)
	})

	r0 = r0.Prev() // we're still pointing to the second element
	r := r0.Link(r1)
	t.Log("r after link:")
	r.Do(func(s core.Source) {
		t.Log(s)
	})

	n = n * 2
	if r.Len() != n {
		t.Fatalf("Unexpected linked r Len: wanted %d, found %d", n, r.Len())
	}

	// ... From godoc:
	// If r and s point to different rings, linking
	// them creates a single ring with the elements of s inserted
	// after r. The result points to the element following the
	// last element of s after insertion.
	//
	if r.Source().ID() != "0" {
		t.Fatalf("Unexpected linked r element: wanted 0, found %s", r.Source().ID())
	}

	for i := 0; i < n; i++ {
		s := fmt.Sprintf("%d", i)
		if r.Source().ID() != s {
			t.Fatalf("%d: Unexpected linked r source ID: wanted %s, found %v", i, s, r.Source().ID())
		}
		r.Next()
	}
}

func TestUnlink(t *testing.T) {
	r := core.NewRing(4)
	for i := 0; i < r.Len(); i++ {
		s := fmt.Sprintf("%d", i)
		r.Set(&mock{s})
		r.Next()
	}

	t.Logf("r ID before unlink: %s", r.Source().ID())
	t.Log("r content before unlink:")
	r.Do(func(s core.Source){
		t.Log(s)
	})


	// Remove second and third element
	s := r.Unlink(2)
	if r.Len() != 2 {
		t.Fatalf("Unexpected unlinked ring Len: wanted 2, found %d", r.Len())
	}

	t.Logf("r ID after unlink: %s", r.Source().ID())
	t.Log("r content after unlink:")
	r.Do(func(s core.Source){
		t.Log(s)
	})

	// We should still point to element 0
	tt := []string{"0", "3"}
	for i, v := range tt {
		if v != r.Source().ID() {
			t.Fatalf("%d: Unexpected source ID: wanted %s, found %s", i, v, r.Source().ID())
		}
		r.Next()
	}

	// Check content of removed subring
	tt = []string{"1", "2"}
	for i, v := range tt {
		if v != s.Source().ID() {
			t.Fatalf("%d: Unexpected source ID: wanted %s, found %s", i, v, s.Source().ID())
		}
		s.Next()
	}

}
