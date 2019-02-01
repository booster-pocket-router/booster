// Copyright © 2019 KIM KeepInMind GmbH/srl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package store_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/booster-proj/booster/core"
	"github.com/booster-proj/booster/store"
)

func TestGet(t *testing.T) {
	s0 := &mock{id: "s0"}
	s1 := &mock{id: "s1"}
	t0 := "foo:port"
	t1 := "bar:port"
	st := &storage{
		index: 0,
		data:  []core.Source{s0, s1},
	}
	s := store.New(st)

	ctx := context.Background()
	src, err := s.Get(ctx, t0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if src.ID() != s0.ID() {
		t.Fatalf("Unexpected source: wanted %s, found %s", s0, src)
	}

	s.AppendPolicy(&store.GenPolicy{
		Name: "p0",
		AcceptFunc: func(id, target string) bool {
			// Does not accept s0 trying to contact t0
			t.Logf("AcceptFunc called with: id(%s) target(%s)", id, target)
			trg := store.TrimPort(t0)
			return !(id == s0.ID() && target == trg)
		},
	})

	src, err = s.Get(ctx, t0)
	if err == nil {
		t.Fatalf("Unexpected source %v, we should have received an error instead", src)
	}

	src, err = s.Get(ctx, t1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if src.ID() != s0.ID() {
		t.Fatalf("Unexpected source: wanted %s, found %s", s0, src)
	}

	st.index = 1 // make storage return s1
	src, err = s.Get(ctx, t0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if src.ID() != s1.ID() {
		t.Fatalf("Unexpected source: wanted %s, found %s", s1, src)
	}
}

func TestMakeBlacklist(t *testing.T) {
	s0 := &mock{id: "s0"}
	s1 := &mock{id: "s1"}
	s2 := &mock{id: "s2"}
	t0 := "t0:port"
	t1 := "t1:port"

	s := store.New(&storage{data: []core.Source{s0, s1, s2}})

	// Without policies the blacklist should be empty
	if bl := s.MakeBlacklist(t0); len(bl) != 0 {
		t.Fatalf("Unexpected blacklist content: wanted [], found %+v", bl)
	}

	p0 := "p0"
	s.AppendPolicy(&store.GenPolicy{
		Name: p0,
		AcceptFunc: func(id, target string) bool {
			// Does not accept s0 trying to contact t0
			t.Logf("AcceptFunc called with: id(%s) target(%s)", id, target)
			trg := store.TrimPort(t0)
			return !(id == s0.ID() && target == trg)
		},
	})

	if bl := s.MakeBlacklist(t0); len(bl) != 1 {
		t.Fatalf("Unexpected blacklist content: wanted [%s], found %+v", s0, bl)
	}
	if bl := s.MakeBlacklist(t1); len(bl) != 0 {
		t.Fatalf("Unexpected blacklist content: wanted [], found %+v", bl)
	}
}

func TestMakeBlacklist_ipify(t *testing.T) {
	en0 := &mock{id: "en0"}
	en4 := &mock{id: "en4"}
	t0 := "50.19.247.198:443"
	t1 := "host.com:443"

	s := store.New(&storage{data: []core.Source{en0, en4}})
	store.Resolver = resolver{
		addrs: []string{
			"50.16.248.221",
			"50.19.247.198",
			"107.22.215.20",
			"54.243.123.39",
			"54.204.36.156",
			"23.21.121.219",
		},
	}
	s.AppendPolicy(store.NewReservedPolicy("T", en0.ID(), "api.ipify.org"))

	if bl := s.MakeBlacklist(t0); len(bl) != 1 {
		t.Fatalf("Unexpected blacklist content: wanted [%s], found %+v", en4, bl)
	}
	if bl := s.MakeBlacklist(t1); len(bl) != 1 {
		t.Fatalf("Unexpected blacklist content: wanted [%s], found %+v", en0, bl)
	}
}

func TestShouldAccept(t *testing.T) {
	s := store.New(&storage{})
	id0 := "foo"
	target := "host:port"

	// Test that the source is accepted if there are no policies.
	if ok, _ := s.ShouldAccept(id0, target); !ok {
		t.Fatalf("Source %s was not accepted, even though it should have", id0)
	}

	// Add a policy that blocks foo
	pid0 := "foo_block"
	s.AppendPolicy(&store.GenPolicy{
		Name: pid0,
		AcceptFunc: func(id, target string) bool {
			return id != "foo"
		},
	})

	ok, p := s.ShouldAccept(id0, target)
	if ok {
		t.Fatalf("Source %s was accepted, even though it shouldn't have", id0)
	}
	if p.ID() != pid0 {
		t.Fatalf("Source %s was correctly blocked, but from the wrong policy: expected %s, found %s", id0, pid0, p.ID())
	}

	// Try with a source that should not be blocked from the
	// last policy added.
	id1 := "bar"
	ok, p = s.ShouldAccept(id1, target)
	if !ok {
		t.Fatalf("Source %s was not accepted, even though it should have", id1)
	}

	// Add a policy that blocks bar
	pid1 := "bar_block"
	s.AppendPolicy(&store.GenPolicy{
		Name: pid1,
		AcceptFunc: func(id, target string) bool {
			return id != "bar"
		},
	})

	ok, p = s.ShouldAccept(id1, target)
	if ok {
		t.Fatalf("Source %s was accepted, even though it shouldn't have", id1)
	}
	if p.ID() != pid1 {
		t.Fatalf("Source %s was correctly blocked, but from the wrong policy: expected %s, found %s", id1, pid1, p.ID())
	}

	// Remove block on bar and check again
	s.DelPolicy(pid1)
	ok, p = s.ShouldAccept(id1, target)
	if !ok {
		t.Fatalf("Source %s was not accepted, even though it should have", id1)
	}
}

func TestAddPolicy(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	if len(s.GetPoliciesSnapshot()) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", s.GetPoliciesSnapshot())
	}

	// Now add a policy.
	s.AppendPolicy(&store.GenPolicy{
		Name: "foo",
		AcceptFunc: func(name, target string) bool {
			return false
		},
	})
	if len(s.GetPoliciesSnapshot()) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", s.GetPoliciesSnapshot())
	}
}

func TestDelPolicy(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	s.AppendPolicy(&store.GenPolicy{
		Name: "foo",
		AcceptFunc: func(name, target string) bool {
			return false
		},
	})
	if len(s.GetPoliciesSnapshot()) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", s.GetPoliciesSnapshot())
	}

	// Now remove the policy.
	s.DelPolicy("foo")
	if len(s.GetPoliciesSnapshot()) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", s.GetPoliciesSnapshot())
	}

}

func TestGetPoliciesSnapshot(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	pl := s.GetPoliciesSnapshot()
	if len(pl) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", pl)
	}

	// Now add a policy.
	s.AppendPolicy(&store.GenPolicy{
		Name: "foo",
		AcceptFunc: func(name, target string) bool {
			return false
		},
	})
	pl = s.GetPoliciesSnapshot()
	if len(pl) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", pl)
	}
}

type mock struct {
	id     string
	active bool
}

func (s *mock) ID() string {
	return s.id
}

func (s *mock) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if s.active {
		return nil, nil
	}
	return nil, fmt.Errorf("no internet connection")
}

func (s *mock) Close() error {
	return nil
}

func (s *mock) String() string {
	return s.ID()
}

type storage struct {
	index int // tells which source should be returned
	data  []core.Source
}

func (s *storage) Put(ss ...core.Source) {
	s.data = append(s.data, ss...)
}

func (s *storage) Del(ss ...core.Source) {
	filtered := make([]core.Source, 0, len(ss))
	filter := func(src core.Source) bool {
		for _, v := range ss {
			if src.ID() == v.ID() {
				return false
			}
		}
		return true
	}
	for _, v := range s.data {
		if filter(v) {
			filtered = append(filtered, v)
		}
	}

	s.data = filtered
}

func (s *storage) Len() int {
	return len(s.data)
}

func (s *storage) Do(f func(core.Source)) {
	for _, v := range s.data {
		f(v)
	}
}

func (s *storage) Get(ctx context.Context, blacklisted ...core.Source) (core.Source, error) {
	isIn := func(s core.Source) bool {
		for _, v := range blacklisted {
			if v.ID() == s.ID() {
				return true
			}
		}
		return false
	}
	src := s.data[s.index]
	if !isIn(src) {
		return src, nil
	}

	return nil, fmt.Errorf("storage: not suitable source found")
}
