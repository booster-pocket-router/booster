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

// Store describes an entity that is able to store,
// delete and enumerate sources.
type Store interface {
	Put(...core.Source)
	Del(...core.Source)

	Len() int
	Do(func(core.Source))
}

type Policy struct {
	Block func(core.Source) bool
	Blocked core.Source
	Reason string
	Code int
}

// A SourceStore is able to keep sources under a set of
// policies, or rules. When it is asked to store a value,
// it performs the policy checks on it, and eventually the
// request is forwarded to the protected store.
type SourceStore struct {
	protected Store
	policies  map[string]Policy
}

func New(store Store) *SourceStore {
	return &SourceStore{
		protected: store,
		policies:  make(map[string]Policy),
	}
}

// GetAccepted returns the list of sources that are actually
// being used by the protected storage. The two lists (the
// complete and the protected one) could differ due to the
// activation of a blocking policy for example.
func (ss *SourceStore) GetAccepted() []core.Source {
	acc := make([]core.Source, 0, ss.protected.Len())
	ss.protected.Do(func(src core.Source) {
		acc = append(acc, src)
	})
	return acc
}

// GetSnapshot returns a copy of the current sources that the store
// is handling. The sources returned are not capable of providing any
// internet connection, but are filled with the policies applied on
// them and the metrics collected.
func (ss *SourceStore) GetSnapshot() []core.Source {
	// TODO: implement
	acc := make([]core.Source, 0, ss.protected.Len())
	ss.protected.Do(func(src core.Source) {
		acc = append(acc, src)
	})
	return acc
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
		if accepted := p.Block(s); !accepted {
			return fmt.Errorf("Policy check: %v", p.Reason)
		}
	}

	// source was accepted by every policy contraint
	return nil
}

func MakeBlockPolicy(id string) Policy {
	return Policy{
		Block: func(s core.Source) bool {
			if s.Name() == id {
				return false
			}
			return true
		},
		Reason: "manually blocked source",
	}
}

func (rs *SourceStore) Put(sources ...core.Source) {
	rs.protected.Put(sources...)
}

func (rs *SourceStore) Del(sources ...core.Source) {
	rs.protected.Del(sources...)
}

func (rs *SourceStore) Len() int {
	return rs.protected.Len()
}

func (rs *SourceStore) Do(f func(core.Source)) {
	rs.protected.Do(f)
}

