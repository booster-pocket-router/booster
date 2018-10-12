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
	"context"
	"net"

	"github.com/booster-proj/core"
)

type srcMock struct {
	id string
}

func (s *srcMock) ID() string {
	return s.id
}

func (s *srcMock) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return nil, nil
}

func (s *srcMock) Metrics() core.Metrics {
	return core.Metrics(make(map[string]interface{}))
}

func TestBalancer_roundRobin(t *testing.T) {
	b := &core.Balancer{}

	if _, err := b.Get(); err == nil {
		t.Fatal("Unexpected nil error with empty balancer")
	}

	s0 := new(srcMock)
	s1 := new(srcMock)
	s0.id = "s0"
	s1.id = "s1"

	b.Put(s0, s1)

	tt := []struct {
		out string
	}{
		{"s0"}, {"s1"}, {"s0"},
	}

	for i, v := range tt {
		// Get sources using the default round robin strategy.
		s, err := b.Get()
		if err != nil {
			t.Fatalf("Unexpected error while getting source: %v. %v", i, err)
		}

		if s.ID() != v.out {
			t.Fatalf("Unexpected source ID: iteration(%v): wanted %v, found %v", i, v.out, s.ID())
		}
	}
}
