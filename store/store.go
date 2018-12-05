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

package store

import (
	"fmt"

	"github.com/booster-proj/booster/core"
)

// A Policy is a simple function that takes a source as
// input and returns wether it should be accepted or not.
type Policy func(core.Source) (bool, error)

type SourceStore struct {
	Balancer core.Balancer
	policies map[string]Policy
}

func (rs *SourceStore) AddPolicy(id string, p Policy) error {
	if rs.policies == nil {
		rs.policies = make(map[string]Policy)
	}

	if _, ok := rs.policies[id]; ok {
		return fmt.Errorf("SourceStore: policy with identifier %s is already present", id)
	}

	rs.policies[id] = p

	// TODO: We need to apply the policy also to the sources
	// that are already in the storage.
	return nil
}

func (rs *SourceStore) DelPolicy(id string) {
	delete(rs.policies, id)
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
