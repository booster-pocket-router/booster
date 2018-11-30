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
	"testing"

	"github.com/booster-proj/booster/source"
)

func TestApplyPolicy_Block(t *testing.T) {
	id := "foo"
	s := &mock{id: id}
	block := source.MakeBlockPolicy(id)

	if err := source.ApplyPolicy(s, block); err == nil {
		t.Fatalf("Source (%v) was accepted, even though it should've been refuted", s)
	}
	s.id = "bar"
	if err := source.ApplyPolicy(s, block); err != nil {
		t.Fatalf("Source (%v) was unexpectedly blocked: %v", s, err)
	}
}
