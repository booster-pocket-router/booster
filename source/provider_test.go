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

package source_test

import (
	"context"
	"testing"
	"time"

	"github.com/booster-proj/booster/source"
	"github.com/booster-proj/booster/core"
)

func TestProvide_cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &source.MergedProvider{}

	c := make(chan error)
	go func() {
		_, err := p.Provide(ctx)
		c <- err
	}()

	cancel()
	select {
	case err := <-c:
		if err != nil && err != ctx.Err() {
			t.Fatalf("Unexpected error: found \"%v\"", err)
		}
	case <-time.After(time.Millisecond * 100):
		t.Fatal("Provide timeout")
	}
}

func TestCheck_cancel(t *testing.T) {
	p := &source.MergedProvider{}
	srcs, _ := p.Provide(context.Background())
	c := make(chan error, len(srcs))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, v := range srcs {
		go func(src core.Source) {
			c <- p.Check(ctx, src, source.High)
		}(v)
	}

	cancel()
	<-time.After(time.Millisecond * 100) // give everyone some time to exit

	for i := 0; i < len(srcs); i++ {
		select {
		case err := <-c:
			if err != nil && err != ctx.Err() {
				t.Fatalf("%d: Unexpected error: %v", i, err)
			}
		default:
			t.Fatalf("%d: Exit dealine exceeded", i)
		}
	}
}
