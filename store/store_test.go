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

type mock struct {
	id     string
	active bool
}

func (s *mock) Name() string {
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

func (s *mock) Value(key interface{}) interface{} {
	return nil
}

func (s *mock) String() string {
	return s.Name()
}

type storage struct {
	data []core.Source
}

func (s *storage) Put(ss ...core.Source) {
	s.data = append(s.data, ss...)
}

func (s *storage) Del(ss ...core.Source) {
	filtered := make([]core.Source, 0, len(ss))
	filter := func(src core.Source) bool {
		for _, v := range ss {
			if src.Name() == v.Name() {
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

func TestAddPolicy(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	if len(s.Policies) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", s.Policies)
	}

	// Now add a policy.
	s.AddPolicy(&store.Policy{
		ID: "block_foo",
		Accept: func(name string) bool {
			return name != "foo"
		},
		Reason: "Some reason",
		Issuer: "Test",
		Code:   500,
	})
	if len(s.Policies) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", s.Policies)
	}
}

func TestDelPolicy(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	s.Policies = append(s.Policies, &store.Policy{
		ID: "block_foo",
		Accept: func(name string) bool {
			return name != "foo"
		},
		Reason: "Some reason",
		Issuer: "Test",
		Code:   500,
	})
	if len(s.Policies) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", s.Policies)
	}

	// Now remove the policy.
	s.DelPolicy("block_foo")
	if len(s.Policies) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", s.Policies)
	}

}

func TestAddDelPolicy_withSideEffects(t *testing.T) {
	// Let's start with a protected storage that contains a
	// source.
	src := &mock{id: "foo"}
	storage := &storage{
		data: []core.Source{src},
	}

	// Create the store.
	s := store.New(storage)

	// Query the store and check that it returns our source
	// (it should not be blocked).
	ss := s.GetProtected()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}

	// Now add a policy that should block our source, and
	// see the results.
	p := &store.Policy{
		ID: "block_foo",
		Accept: func(name string) bool {
			return name != "foo"
		},
		Reason: "Some reason",
		Issuer: "Test",
		Code:   500,
	}
	s.AddPolicy(p)

	// Now check if the source is actually blocked.
	ss = s.GetProtected()
	if len(ss) != 0 {
		t.Fatalf("Unexpected accepted sources: wanted len == 0, found %+v", ss)
	}

	// Remove the policies and check the result again.
	s.DelPolicy(p.ID)
	ss = s.GetProtected()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}
}

func TestPut(t *testing.T) {
	// Build the protected storage.
	storage := &storage{
		data: []core.Source{},
	}
	// Create the store
	s := store.New(storage)

	// Test that it is actually possible to Put a source.
	s0 := &mock{id: "foo"}
	s.Put(s0)

	ss := s.GetProtected()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}

	// Now add a blocking policy and check wether we're able
	// to Put sources or not.
	p := &store.Policy{
		ID: "block_bar",
		Accept: func(name string) bool {
			return name != "bar"
		},
		Reason: "Some reason",
		Issuer: "Test",
		Code:   500,
	}
	if err := s.AddPolicy(p); err != nil {
		t.Fatalf("Unexpected error while adding policy %v: %v", p, err)
	}

	s1 := &mock{id: "bar"}
	s.Put(s1)

	ss = s.GetProtected()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}

	// Ensure that it is not possible to add the same policy multiple times
	if err := s.AddPolicy(p); err == nil {
		t.Fatalf("We were allowed to add the same policy twice, but we shouldn't")
	}

	// If the policy is removed, the source should be eventually integrated
	// into the accepted sources.
	s.DelPolicy(p.ID)

	ss = s.GetProtected()
	if len(ss) != 2 {
		t.Fatalf("Unexpected accepted sources: wanted len == 2, found: %+v", ss)
	}
}

func TestDel(t *testing.T) {
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}

	// Build the protected storage with a source in it.
	storage := &storage{
		data: []core.Source{s0},
	}
	// Create the store.
	s := store.New(storage)

	// Now add a blocking policy and put a source into the under
	// policy limbo.
	p := &store.Policy{
		ID: "block_bar",
		Accept: func(name string) bool {
			return name != "bar"
		},
		Reason: "Some reason",
		Issuer: "Test",
		Code:   500,
	}
	s.AddPolicy(p)
	s.Put(s1) // blocked by the policy

	ss := s.GetProtected()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}

	// Now delete without removing the policy (otherwise s1 will
	// be inserted into the same storage as s0)
	s.Del(s0, s1)
	ss = s.GetProtected()
	if len(ss) != 0 {
		t.Fatalf("Unexpected accepted sources: wanted len == 0, found: %+v", ss)
	}

	// Remove the policy: no sources should added to the storage.
	s.DelPolicy(p.ID)
	ss = s.GetProtected()
	if len(ss) != 0 {
		t.Fatalf("Unexpected accepted sources: wanted len == 0, found: %+v", ss)
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
	s.AddPolicy(&store.Policy{
		ID: "block_foo",
		Accept: func(name string) bool {
			return name != "foo"
		},
		Reason: "Some reason",
		Issuer: "Test",
		Code:   500,
	})
	pl = s.GetPoliciesSnapshot()
	if len(pl) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", pl)
	}
}
