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

package source_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/booster-proj/booster/core"
	"github.com/booster-proj/booster/source"
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

func TestHooker(t *testing.T) {
	h := &source.Hooker{}
	ref := "foo"
	h.HandleDialErr(ref, "net", "addr", errors.New("some error"))

	if err := h.HookErr(ref); err == nil {
		t.Fatalf("Wanted hook error for id %s, found nil", ref)
	}
	if err := h.HookErr(ref); err != nil {
		t.Fatalf("Wanted nil error for id %s, found %v", ref, err)
	}
}

type storage struct {
	data    []core.Source
	putHook func(ss ...core.Source)
	delHook func(ss ...core.Source)
}

func (s *storage) Put(ss ...core.Source) {
	s.data = append(s.data, ss...)
	if f := s.putHook; f != nil {
		f(ss...)
	}
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
	if f := s.delHook; f != nil {
		f(ss...)
	}
}

func (s *storage) GetAccepted() []core.Source {
	return s.data
}

func (s *storage) Len() int {
	return len(s.data)
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

func (p *mockProvider) Check(ctx context.Context, src core.Source, level source.Confidence) error {
	switch level {
	case source.High:
		_, err := src.DialContext(ctx, "net", "addr")
		return err
	default:
		return nil
	}
}

func TestRun_cancel(t *testing.T) {
	s := new(storage)
	l := source.NewListener(s)
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

func mocksFrom(s ...string) []core.Source {
	ret := make([]core.Source, len(s))
	for i, v := range s {
		ret[i] = &mock{id: v}
	}
	return ret
}

func sameContent(a, b []core.Source) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := make([]core.Source, len(a))
	sortedB := make([]core.Source, len(b))
	copy(sortedA, a)
	copy(sortedB, b)

	sort.SliceStable(sortedA, func(i, j int) bool { return sortedA[i].Name() > sortedA[j].Name() })
	sort.SliceStable(sortedB, func(i, j int) bool { return sortedB[i].Name() > sortedB[j].Name() })

	for i, v := range sortedA {
		if v.Name() != sortedB[i].Name() {
			return false
		}
	}
	return true
}

func TestDiff(t *testing.T) {
	tt := []struct {
		old    []core.Source
		cur    []core.Source
		add    []core.Source
		remove []core.Source
	}{
		{old: mocksFrom("1"), cur: mocksFrom("1", "2"), add: mocksFrom("2"), remove: mocksFrom()},
		{old: mocksFrom("1", "2"), cur: mocksFrom("2"), add: mocksFrom(), remove: mocksFrom("1")},
		{old: mocksFrom("1", "2"), cur: mocksFrom("1", "2"), add: mocksFrom(), remove: mocksFrom()},
		{old: mocksFrom("1", "2"), cur: mocksFrom("3", "4"), add: mocksFrom("3", "4"), remove: mocksFrom("1", "2")},
	}

	for i, v := range tt {
		add, remove := source.Diff(v.old, v.cur)
		if !sameContent(add, v.add) {
			t.Fatalf("%d: Unexpected add context: wanted %v, found %v", i, v.add, add)
		}
		if !sameContent(remove, v.remove) {
			t.Fatalf("%d: Unexpected remove context: wanted %v, found %v", i, v.remove, remove)
		}

		ccur := make([]core.Source, len(v.old), cap(v.cur)+cap(v.old))
		copy(ccur, v.old)           // add old content
		ccur = append(ccur, add...) // add things

		// remove things
		f := func(src core.Source) bool {
			for _, v := range remove {
				if v.Name() == src.Name() {
					return false
				}
			}
			return true
		}

		fccur := ccur[:0]
		for _, v := range ccur {
			if f(v) {
				fccur = append(fccur, v)
			}
		}

		if !sameContent(fccur, v.cur) {
			t.Fatalf("%d: Unexpected final content: wanted %v, found %v", i, v.cur, fccur)
		}
	}
}

func TestPoll(t *testing.T) {
	putc := make(chan core.Source, 1)
	delc := make(chan core.Source, 1)
	s := &storage{
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
	en0 := &mock{id: "en0", active: true}
	awl0 := &mock{id: "awl0", active: false}
	p := &mockProvider{
		sources: []*mock{en0, awl0},
	}
	l := source.NewListener(s)
	l.Provider = p

	ctx := context.Background()
	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{en0.Name()} {
		select {
		case s := <-putc:
			if v != s.Name() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.Name())
			}
		case <-time.After(time.Millisecond * 200):
			t.Fatalf("%d: Deadline exceeded", i)
		}
	}

	awl0.active = true // set awl0 to active
	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{awl0.Name()} {
		select {
		case s := <-putc:
			if v != s.Name() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.Name())
			}
		case <-time.After(time.Millisecond * 100):
		}
	}

	p.sources = p.sources[:1]
	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{awl0.Name()} {
		select {
		case s := <-delc:
			if v != s.Name() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.Name())
			}
		case <-time.After(time.Millisecond * 200):
			t.Fatalf("%d: Deadline exceeded", i)
		}
	}

	// Add awl0 again, this time with an active internet
	// connection.
	p.sources = append(p.sources, awl0)

	if err := l.Poll(ctx); err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{awl0.Name()} {
		select {
		case s := <-putc:
			if v != s.Name() {
				t.Fatalf("%d: Unexpected source id: wanted %s, found %s", i, v, s.Name())
			}
		case <-time.After(time.Millisecond * 200):
			t.Fatalf("%d: Deadline exceeded", i)
		}
	}
}
