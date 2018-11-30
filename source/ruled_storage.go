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

package source

import (
	"fmt"
	"github.com/booster-proj/booster/core"
)

// A Policy is a simple function that takes a source as
// input and returns wether it should be accepted or not.
type Policy func(core.Source) (bool, error)

type RuledStorage struct {
	Storage
	policies map[string]Policy
}

func NewRuledStorage(s Storage) *RuledStorage {
	return &RuledStorage{Storage: s, policies: make(map[string]Policy)}
}

func (rs *RuledStorage) AddPolicy(id string, p Policy) error {
	if rs.policies == nil {
		rs.policies = make(map[string]Policy)
	}

	if _, ok := rs.policies[id]; ok {
		return fmt.Errorf("RuledStorage: policy with identifier %s is already present", id)
	}

	rs.policies[id] = p

	// TODO: We need to apply the policy also to the sources
	// that are already in the storage.
	return nil
}

func ApplyPolicy(s core.Source, policies ...Policy) error {
	for _, p := range policies {
		if accepted, err := p(s); !accepted {
			return fmt.Errorf("Policy check: %v", err)
		}
	}

	// source was accepted by every policy contraint
	return nil
}

func MakeBlockPolicy(id string) Policy {
	return func(s core.Source) (bool, error) {
		if s.ID() == id {
			return false, fmt.Errorf("source %s is blocked", s.ID())
		}
		return true, nil
	}
}
