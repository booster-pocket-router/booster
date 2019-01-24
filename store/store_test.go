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

func TestAddPolicy(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	if len(s.Policies) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", s.Policies)
	}

	// Now add a policy.
	s.AppendPolicy(&store.Policy{
		ID: "foo",
		Accept: func(name, target string) bool {
			return false
		},
		Reason: "undefined",
		Issuer: "testing.T",
		Code:   -1,
	})
	if len(s.Policies) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", s.Policies)
	}
}

func TestDelPolicy(t *testing.T) {
	s := store.New(&storage{
		data: []core.Source{},
	})
	s.AppendPolicy(&store.Policy{
		ID: "foo",
		Accept: func(name, target string) bool {
			return false
		},
		Reason: "undefined",
		Issuer: "testing.T",
		Code:   -1,
	})
	if len(s.Policies) != 1 {
		t.Fatalf("Unexpected policies count: wanted 1, found %+v", s.Policies)
	}

	// Now remove the policy.
	s.DelPolicy("foo")
	if len(s.Policies) != 0 {
		t.Fatalf("Unexpected policies count: wanted 0, found %+v", s.Policies)
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
	s.AppendPolicy(&store.Policy{
		ID: "foo",
		Accept: func(name, target string) bool {
			return false
		},
		Reason: "undefined",
		Issuer: "testing.T",
		Code:   -1,
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
	data []core.Source
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

	for _, v := range s.data {
		if !isIn(v) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("storage: not suitable source found")
}
