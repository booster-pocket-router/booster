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

func TestApplyPolicy_Block(t *testing.T) {
	id := "foo"
	s := &mock{id: id}
	block := store.MakeBlockPolicy(id)

	if err := store.ApplyPolicy(s, block); err == nil {
		t.Fatalf("Source (%v) was accepted, even though it should've been refuted", s)
	}
	s.id = "bar"
	if err := store.ApplyPolicy(s, block); err != nil {
		t.Fatalf("Source (%v) was unexpectedly blocked: %v", s, err)
	}
}
