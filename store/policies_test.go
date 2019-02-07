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
	"testing"

	"github.com/booster-proj/booster/store"
)

type resolver struct {
	// addresses to return
	addrs []string
	host  string
}

func (r resolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	if len(r.addrs) > 0 {
		return r.addrs, nil
	}

	return []string{host}, nil
}

func (r resolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	return []string{r.host}, nil
}

func TestTrimPort(t *testing.T) {
	tt := []struct {
		in  string
		out string
	}{
		{in: "foo:port", out: "foo"},
		{in: "example.com:443", out: "example.com"},
	}

	for i, v := range tt {
		addr := store.TrimPort(v.in)
		if addr != v.out {
			t.Fatalf("%d: unexpected address: wanted %s, found %s", i, v.out, addr)
		}
	}
}

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
	store.Resolver = resolver{}
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}
	t0 := "host0"
	t1 := "host1"
	t2 := "host2"

	p := store.NewReservedPolicy("T", s0.ID(), t0)
	if ok := p.Accept(s0.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t1); ok {
		t.Fatalf("Policy %s accepted source %v for address %s", p.ID(), s1.ID(), t1)
	}
	if ok := p.Accept(s1.ID(), t0); ok {
		t.Fatalf("Policy %s accepted source %v for address %s", p.ID(), s1.ID(), t0)
	}
	if ok := p.Accept(s1.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t1)
	}

	// reserved policy with multiple addresses
	p = store.NewReservedPolicy("T", s0.ID(), t0, t1)
	if ok := p.Accept(s0.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t2); ok {
		t.Fatalf("Policy %s accepted source %v for address %s", p.ID(), s0.ID(), t2)
	}
}

func TestAvoidPolicy(t *testing.T) {
	store.Resolver = resolver{}
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}
	t0 := "host0"
	t1 := "host1"

	p := store.NewAvoidPolicy("T", s0.ID(), t0)
	if ok := p.Accept(s0.ID(), t0); ok {
		t.Fatalf("Policy %s accepted source %v for address %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t1)
	}
	if ok := p.Accept(s1.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t0)
	}
	if ok := p.Accept(s1.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t1)
	}
}

func TestStickyPolicy(t *testing.T) {
	store.Resolver = resolver{}
	s0 := &mock{id: "foo"}
	s1 := &mock{id: "bar"}
	t0 := "host0"
	t1 := "host1"

	history := make(map[string]string)
	p := store.NewStickyPolicy("T", func(address string) (src string, ok bool) {
		src, ok = history[address]
		return
	})

	if ok := p.Accept(s0.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s1.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s0.ID(), t1)
	}
	if ok := p.Accept(s1.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t1)
	}

	history[t0] = s0.ID()
	if ok := p.Accept(s0.ID(), t0); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s0.ID(), t0)
	}
	if ok := p.Accept(s1.ID(), t0); ok {
		t.Fatalf("Policy %s accepted source %v for address %s", p.ID(), s1.ID(), t0)
	}
	if ok := p.Accept(s0.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s0.ID(), t1)
	}
	if ok := p.Accept(s1.ID(), t1); !ok {
		t.Fatalf("Policy %s did not accept source %v for address %s", p.ID(), s1.ID(), t1)
	}
}
