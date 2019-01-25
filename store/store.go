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

package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/booster-proj/booster/core"
)

// Store describes an entity that is able to store,
// delete and enumerate sources.
type Store interface {
	Put(...core.Source)
	Del(...core.Source)
	Get(context.Context, ...core.Source) (core.Source, error)

	Len() int
	Do(func(core.Source))
}

// A Policy defines wether a connection to `target` should
// be accepted by source `id`.
type Policy interface {
	ID() string
	Accept(id, target string) bool
}

// A SourceStore is able to keep sources under a set of
// policies, or rules. When it is asked to store a value,
// it performs the policy checks on it, and eventually the
// request is forwarded to the protected store.
type SourceStore struct {
	protected Store

	mux      sync.Mutex
	Policies []Policy
}

// DummySource is a representation of a source, suitable
// when other components need information about the sources stored,
// but should not be able to mess with it's actual content.
type DummySource struct {
	ID string `json:"name"`
}

// New creates a New instance of SourceStore, using interally `store`
// as the protected storage.
func New(store Store) *SourceStore {
	return &SourceStore{
		protected: store,
		Policies:  []Policy{},
	}
}

// Get is an implementation of booster.Balancer. It provides a source, avoiding
// the ones `blacklisted`. The `blacklisted` list is populated with the sources
// that cannot be accepted due to policy restrictions. The source is then
// retriven from the protected storage.
func (ss *SourceStore) Get(ctx context.Context, target string, blacklisted ...core.Source) (core.Source, error) {
	blacklisted = append(blacklisted, ss.MakeBlacklist(target)...)
	return ss.protected.Get(ctx, blacklisted...)
}

// ShouldAccept takes `id` and `target`, iterates through the list of policies
// and returns false if the two inputs are not accepted by one of them. The
// offending policy is also returned.
// Returns true if no policy blocks `id` and `target`.
func (ss *SourceStore) ShouldAccept(id, target string) (bool, Policy) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	for _, p := range ss.Policies {
		if ok := p.Accept(id, target); !ok {
			return ok, p
		}
	}

	return true, nil
}

// MakeBlacklist computes the list of blacklisted sources for `target`, i.e. the
// sources that should not be used to perform a request to `target`, because there
// is one or more policies that do not accept them.
func (ss *SourceStore) MakeBlacklist(target string) []core.Source {
	acc := make([]core.Source, 0, ss.Len())

	// return immediately if there is no policy.
	ss.mux.Lock()
	l := len(ss.Policies)
	ss.mux.Unlock()
	if l == 0 {
		return acc
	}

	ss.Do(func(src core.Source) {
		if ok, _ := ss.ShouldAccept(src.ID(), target); !ok {
			acc = append(acc, src)
		}
	})

	return acc
}

// Len returns the number of sources available to the store.
func (ss *SourceStore) Len() int {
	return ss.protected.Len()
}

// Do executes `f` on each source of the protected storage.
func (ss *SourceStore) Do(f func(core.Source)) {
	ss.protected.Do(f)
}

// AppendPolicy appends `p` to the end of the list of sources. Remember
// that the policyes are always applied in order, and the chain of checks
// is interrupted as soon as a policy refutes to accept a source, i.e. the
// policies that come after that are not executed.
func (ss *SourceStore) AppendPolicy(p Policy) error {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	if ss.Policies == nil {
		ss.Policies = make([]Policy, 0, 1)
	}

	// Ensure that this is not a duplicate.
	for _, v := range ss.Policies {
		if v.ID() == p.ID() {
			return fmt.Errorf("source store: a policy with identifier %v is already present", v.ID())
		}
	}

	// Eventually append the new policy.
	ss.Policies = append(ss.Policies, p)

	return nil
}

// DelPolicy removes the policy with identifier `id` from the storage.
func (ss *SourceStore) DelPolicy(id string) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	// Remove the policy from the storage.
	var j int
	var found bool
	for i, v := range ss.Policies {
		if v.ID() == id {
			found = true
			j = i
			break
		}
	}
	if !found {
		return
	}
	// avoid any possible memory leak in the underlying array.
	ss.Policies[j] = nil
	ss.Policies = append(ss.Policies[:j], ss.Policies[j+1:]...)
}

// Put adds `sources` to the protected storage.
func (ss *SourceStore) Put(sources ...core.Source) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	ss.protected.Put(sources...)
}

// Del removes `sources` from the protected storage.
func (ss *SourceStore) Del(sources ...core.Source) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	ss.protected.Del(sources...)
}

// GetPoliciesSnapshot returns a copy of the current policies
// active in the store.
func (ss *SourceStore) GetPoliciesSnapshot() []Policy {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	acc := make([]Policy, len(ss.Policies))
	copy(acc, ss.Policies)
	return acc
}

// GetSourcesSnapshot returns nothing more then a copy of the
// list of sources that the storage is holding.
func (ss *SourceStore) GetSourcesSnapshot() []*DummySource {
	acc := make([]*DummySource, 0, ss.protected.Len())

	ss.protected.Do(func(src core.Source) {
		acc = append(acc, &DummySource{
			ID: src.ID(),
		})
	})

	return acc
}
