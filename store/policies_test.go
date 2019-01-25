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
	"github.com/booster-proj/booster/store"
	"testing"
)

func TestBlockPolicy(t *testing.T) {
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}
	p := store.NewBlockPolicy("T", "foo")

	if ok := p.Accept(s1.ID(), ""); !ok {
		t.Fatalf("Policy %s did not accept source %v", p.ID(), s1.ID())
	}
	if ok := p.Accept(s0.ID(), ""); ok {
		t.Fatalf("Policy %s accepted source %v", p.ID(), s0.ID())
	}
}

func TestReservedPolicy(t *testing.T) {
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}
	t0 := "host0:port"
	t1 := "host1:port"

	p := store.NewReservedPolicy("T", s0.ID(), t0)
	if ok := p.Accept(s0.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for target %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t1); ok {
		t.Fatalf("Policy %s accepted source %v for target %s", p.ID(), s1.ID(), t1)
	}
	if ok := p.Accept(s1.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for target %s", p.ID(), s1.ID(), t0)
	}
	if ok := p.Accept(s1.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for target %s", p.ID(), s1.ID(), t1)
	}
}

func TestAvoidPolicy(t *testing.T) {
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}
	t0 := "host0:port"
	t1 := "host1:port"

	p := store.NewAvoidPolicy("T", s0.ID(), t0)
	if ok := p.Accept(s0.ID(), t0); ok {
		t.Fatalf("Policy %s accepted source %v for target %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for target %s", p.ID(), s1.ID(), t1)
	}
	if ok := p.Accept(s1.ID(), t0); !ok {
		t.Fatalf("Policy %s did not source %v for target %s", p.ID(), s1.ID(), t0)
	}
	if ok := p.Accept(s1.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for target %s", p.ID(), s1.ID(), t1)
	}
}
