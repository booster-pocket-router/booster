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

package listener_test

import (
	"context"
	"net"
	"fmt"
	"testing"
	"time"

	"github.com/booster-proj/booster/listener"
	"github.com/booster-proj/booster/listener/provider"
	"github.com/booster-proj/core"
)

type mock struct {
	id string
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

func (s *mock) Metrics() map[string]interface{} {
	return make(map[string]interface{})
}

type storage struct {
	putHook func(ss ...core.Source)
	delHook func(ss ...core.Source)
}

func (s *storage) Put(ss ...core.Source) {
	if f := s.putHook; f != nil {
		f(ss...)
	}
}

func (s *storage) Del(ss ...core.Source) {
	if f := s.delHook; f != nil {
		f(ss...)
	}
}

type mockProvider struct {
	sources []*mock
}

func (p *mockProvider) Provide(ctx context.Context) ([]core.Source, error) {
	list := make([]core.Source, len(p.sources))
	for i, v := range p.sources {
		list[i] = v
	}
	return list, nil
}

func (p *mockProvider) Check(ctx context.Context, src core.Source, level provider.Confidence) error {
	switch level {
	case provider.Low:
		return nil
	case provider.High:
		_, err := src.DialContext(ctx, "net", "addr")
		return err
	}
	return nil
}

func TestRun_cancel(t *testing.T) {
	s := new(storage)
	l := listener.New(s)
	c := make(chan error)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go func() {
		c <- l.Run(ctx)
	}()

	cancel()
	select {
	case err := <-c:
		if err != ctx.Err() {
			t.Fatalf("Unexpected Run error: wanted %v, found %v", ctx.Err(), err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Run took too long to return")
	}

}

func TestPoll(t *testing.T) {
	putc := make(chan core.Source, 1)
	delc := make(chan core.Source, 1)
	s := &storage {
		putHook: func(ss ...core.Source) {
			go func() {
				for _, v := range ss {
					putc <- v
				}
			}()
		},
		delHook: func(ss ...core.Source) {
			go func() {
				for _, v := range ss {
					delc <- v
				}
			}()
		},
	}
	p := &mockProvider{
		sources: []*mock{&mock{id: "en0", active: true}, &mock{id: "awl0", active: false}},
	}
	l := listener.New(s)
	l.Provider = p

	ctx := context.Background()
	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{"en0"} {
		select {
		case s := <-putc:
			if v != s.ID() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.ID())
			}
		case <-time.After(time.Millisecond*200):
			t.Fatalf("%d: Deadline exceeded", i)
		}
	}

	p.sources[1].active = true // set awl0 to active
	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	// even though the source is now active, we should not
	// be able to detect it until we remove & add the src
	// again.
	for i, _ := range []string{""} {
		select {
		case s := <-putc:
			t.Fatalf("%d: Unexpected source id: wanted empty, found %s", i, s.ID())
		case <-time.After(time.Millisecond*100):
		}
	}

	p.sources = p.sources[:1]
	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{"awl0"} {
		select {
		case s := <-delc:
			if v != s.ID() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.ID())
			}
		case <-time.After(time.Millisecond*200):
			t.Fatalf("%d: Deadline exceeded", i)
		}
	}

	// Add awl0 again, this time with an active internet
	// connection.
	p.sources = append(p.sources, &mock{id: "awl0", active: true})

	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{"awl0"} {
		select {
		case s := <-putc:
			if v != s.ID() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.ID())
			}
		case <-time.After(time.Millisecond*200):
			t.Fatalf("%d: Deadline exceeded", i)
		}
	}

}
