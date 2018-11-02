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
	"testing"
	"time"

	"github.com/booster-proj/booster/listener"
	"github.com/booster-proj/core"
)

type mock struct {
	id string
}

func (s *mock) ID() string {
	return s.id
}

func (s *mock) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return nil, nil
}

func (s *mock) Metrics() map[string]interface{} {
	return make(map[string]interface{})
}

type storage struct {
}

func (s *storage) Put(ss ...core.Source) {
}

func (s *storage) Del(ss ...core.Source) {
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
