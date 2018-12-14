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
	"net"
	"testing"

	"github.com/booster-proj/booster/source"
)

func TestFollow(t *testing.T) {
	conn0, _ := net.Pipe()

	iti0 := &source.Interface{}

	l := iti0.Len()
	if l != 0 {
		t.Fatalf("Unexpected Len: wanted 0, found %d", l)
	}

	if _, err := iti0.Follow(conn0); err != nil {
		t.Fatal(err)
	}
	l = iti0.Len()
	if l != 1 {
		t.Fatalf("Unexpected Len: wanted 1, found %d", l)
	}

	if err := iti0.Close(); err != nil {
		t.Fatal(err)
	}
	l = iti0.Len()
	if l != 0 {
		t.Fatalf("Unexpected Len: wanted 0, found %d", l)
	}
}