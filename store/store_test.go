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

func TestAddDelPolicy(t *testing.T) {
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
	ss := s.GetAccepted()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}

	// Now add a policy that should block our source, and
	// see the results.
	p := &store.Policy{
		ID: "bar",
		Func: func(name string) bool {
			return name != "foo"
		},
		Reason: "Some reason",
		Code:   500,
	}
	s.AddPolicy(p)

	// Now check if the source is actually blocked.
	ss = s.GetAccepted()
	if len(ss) != 0 {
		t.Fatalf("Unexpected accepted sources: wanted len == 0, found %+v", ss)
	}

	// Remove the policies and check the result again.
	s.DelPolicy(p.ID)
	ss = s.GetAccepted()
	if len(ss) != 1 {
		t.Fatalf("Unexpected accepted sources: wanted len == 1, found: %+v", ss)
	}
}
