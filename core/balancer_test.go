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

package core_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/booster-proj/booster/core"
)

type mock struct {
	id        string
	closeHook func()
}

func newMock(id string) *mock {
	return &mock{id: id}
}

func (s *mock) ID() string {
	return s.id
}

func (s *mock) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return nil, nil
}

func (s *mock) Close() error {
	if f := s.closeHook; f != nil {
		go f()
	}
	return nil
}

// Just test that Len does not panic.
func TestLen(t *testing.T) {
	b := &core.Balancer{}
	if b.Len() != 0 {
		t.Fatalf("Unexpected balancer Len: wanted 0, found %d", b.Len())
	}
}

// Just test that put does not panic.
func TestPut(t *testing.T) {
	b := &core.Balancer{}
	s := newMock("s0")

	t.Logf("Put %v into balancer(size: %d)", s, b.Len())
	b.Put(s)

	if b.Len() != 1 {
		t.Fatalf("Unexpected balancer Len: wanted 1, found %d", b.Len())
	}

	b.Do(func(s core.Source) {
		if s.ID() != "s0" {
			t.Fatalf("Unexpected source Identifier: wanted s0, found %s", s.ID())
		}
	})
}

func TestPut_empty(t *testing.T) {
	b := &core.Balancer{}
	s := newMock("s0")

	t.Logf("Put %v into balancer(size: %d, value: %+v)", s, b.Len(), b)
	b.Put(s)
	b.Del(s)

	t.Logf("Put %v into balancer(size: %d, value: %+v)", s, b.Len(), b)
	b.Put(s)
}

// Test balancer with its default round robin strategy.
func TestGet_roundRobin(t *testing.T) {
	b := &core.Balancer{}

	if _, err := b.Get(context.TODO()); err == nil {
		t.Fatal("Unexpected nil error with empty balancer")
	}

	s0 := newMock("s0")
	s1 := newMock("s1")
	s2 := newMock("s2")

	b.Put(s0, s1, s2)

	tt := []struct {
		out string
	}{
		{"s0"}, {"s1"}, {"s2"}, {"s0"},
	}

	for i, v := range tt {
		// Get sources using the default round robin strategy.
		s, err := b.Get(context.TODO())
		if err != nil {
			t.Fatalf("Unexpected error while getting source: %v. %v", i, err)
		}

		if s.ID() != v.out {
			t.Fatalf("Unexpected source ID: iteration(%v): wanted %v, found %v", i, v.out, s.ID())
		}
	}
}

func TestGetBlacklist_roundRobin(t *testing.T) {
	b := &core.Balancer{}

	s0 := newMock("s0")
	s1 := newMock("s1")

	b.Put(s0, s1)

	s, _ := b.Get(context.TODO())
	if s.ID() != "s0" {
		t.Fatalf("Unexpected source ID: wanted %v, found %v", s0.ID(), s1.ID())
	}

	s2, err := b.Get(context.TODO(), s1)
	if err != nil {
		t.Fatalf("Unexpected error while getting source: %v", err)
	}
	if s2.ID() != "s0" {
		t.Fatalf("Unexpected source ID: wanted %v, found %v, blacklisted source %v", s0.ID(), s2.ID(), s1.ID())
	}
}

func TestDel(t *testing.T) {
	b := &core.Balancer{}

	s0 := newMock("s0")
	s1 := newMock("s1")

	b.Put(s0, s1)

	n := b.Len()
	t.Logf("Inserted %v elements into previously emtpy balancer", n)

	c := make(chan bool, 1)
	s0.closeHook = func() {
		t.Log("closeHook() called")
		c <- true
	}

	b.Del(s0)

	n = n - 1
	if b.Len() != n {
		t.Fatalf("Unexpected balancer Len after Del: wanted %v, found %v", n, b.Len())
	}

	b.Do(func(s core.Source) {
		if s.ID() != "s1" {
			t.Fatalf("Unexpected source ID: wanted s1, found %v", s.ID())
		}
	})

	select {
	case <-c:
		return
	case <-time.After(time.Millisecond):
		t.Fatal("closeHook was not called")
	}
}
