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
	"fmt"
	"sync"

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

const (
	PolicyBlock int = 101
)

type Policy struct {
	// ID is used to identify later a policy.
	ID string `json:"id"`
	// Accept is the function used to check wether this policy
	// is applied to item with name == name or not. Returns
	// true if the input should be blocked/not accepted.
	Accept func(name string) bool `json:"-"`
	// Reason explains why this policy exists.
	Reason string `json:"reason"`
	// Issuer tells where this policy comes from.
	Issuer string `json:"issuer"`
	// Code is the code of the policy, usefull when the policy
	// is delivered to another context.
	Code int `json:"code"`
}

func (p *Policy) String() string {
	return p.ID
}

// A SourceStore is able to keep sources under a set of
// policies, or rules. When it is asked to store a value,
// it performs the policy checks on it, and eventually the
// request is forwarded to the protected store.
type SourceStore struct {
	protected Store

	mux         sync.Mutex
	Policies    []*Policy
	underPolicy []*DummySource
}

// A DummySource is a source which stores only the information
// of it's parent source at copy time, but it is no longer able
// to produce any internet connection. It should be used to show // snapshots of the current storage to other componets of the
// program that should not be able to break or work with the
// original and active source.
type DummySource struct {
	internal core.Source `json:"-"`
	Name     string      `json:"name"`
	Policy   *Policy     `json:"policy"`
	Blocked  bool        `json:"blocked"`
}

func New(store Store) *SourceStore {
	return &SourceStore{
		protected:   store,
		Policies:    []*Policy{},
		underPolicy: []*DummySource{},
	}
}

// GetAccepeted returns the list sources stored in the
// protected storage.
func (ss *SourceStore) GetProtected() []core.Source {
	acc := make([]core.Source, 0, ss.protected.Len())
	ss.protected.Do(func(src core.Source) {
		acc = append(acc, src)
	})
	return acc
}

// GetActive returns the list of protected source plus the
// list of sources active, but blocked by a policy.
func (ss *SourceStore) GetActive() []core.Source {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	// Append protected sources...
	acc := make([]core.Source, 0, ss.protected.Len()+len(ss.underPolicy))
	ss.protected.Do(func(src core.Source) {
		acc = append(acc, src)
	})

	// ...and the ones under policy.
	for _, v := range ss.underPolicy {
		if v.internal != nil {
			acc = append(acc, v.internal)
		}
	}

	return acc
}

// GetSourcesSnapshot returns a copy of the current sources that the store
// is handling. The sources returned are not capable of providing any
// internet connection, but are filled with the policies applied on
// them and the metrics collected.
func (ss *SourceStore) GetSourcesSnapshot() []*DummySource {
	acc := make([]*DummySource, 0, ss.protected.Len()+len(ss.underPolicy))

	ss.protected.Do(func(src core.Source) {
		acc = append(acc, &DummySource{
			Name:    src.Name(),
			Blocked: false,
		})
	})

	for _, v := range ss.underPolicy {
		acc = append(acc, &DummySource{
			Name:    v.Name,
			Blocked: v.Blocked,
			Policy:  v.Policy,
		})
	}

	return acc
}

// Add policy stores the policy and applies it also to the sources
// stored in the protected storage, removing them from it if
// required.
func (ss *SourceStore) AddPolicy(p *Policy) error {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	if ss.Policies == nil {
		ss.Policies = make([]*Policy, 0, 1)
	}

	// Ensure that this is not a duplicate
	for _, v := range ss.Policies {
		if v.ID == p.ID {
			return fmt.Errorf("source store: a policy with identifier %v is already present", v.ID)
		}
	}

	// Start the integration process
	ss.Policies = append(ss.Policies, p)

	// Now apply the new policy to the items that
	// are already in the storage.
	acc := make([]core.Source, 0, ss.protected.Len())
	ss.protected.Do(func(src core.Source) {
		if !p.Accept(src.Name()) {
			// the source was not accepted by
			// the policy.
			acc = append(acc, src)
		}
	})

	// Remove the unaccepted sources from the protected
	// storage.
	ss.protected.Del(acc...)

	// Now keep a trace of the sources that are under
	// policy.
	if ss.underPolicy == nil {
		ss.underPolicy = make([]*DummySource, 0, len(acc))
	}
	for _, v := range acc {
		ss.underPolicy = append(ss.underPolicy, &DummySource{
			internal: v,
			Name:     v.Name(),
			Blocked:  true,
			Policy:   p,
		})
	}

	return nil
}

// DelPolicy removes the policy with identifier id from the storage.
// It then loops through each source under policy, and frees it if
// the policy is the removed one, putting the source again in the
// protected storage.
// Note that only the first instance of policy with identifier id is
// removed.
func (ss *SourceStore) DelPolicy(id string) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	// Remove the policy from the storage.
	var j int
	var found bool
	for i, v := range ss.Policies {
		if v.ID == id {
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

	// Now restore the sources under policy.
	if ss.underPolicy == nil {
		return
	}

	acc := make([]*DummySource, 0, len(ss.underPolicy))
	for _, v := range ss.underPolicy {
		if v.Policy.ID == id {
			// Restore this source!
			ss.protected.Put(v.internal)
		} else {
			acc = append(acc, v)
		}
	}
	ss.underPolicy = acc
}

// Put adds sources to the protected storage, if allowed
// by the policies stored. Otherwise the source is added to
// a  temporary storage of sources under policy, and
// eventually put into the protected storage if the blocking
// policy is removed.
func (ss *SourceStore) Put(sources ...core.Source) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	sf := func(src core.Source) (*Policy, bool) {
		for _, v := range ss.Policies {
			if !v.Accept(src.Name()) {
				return v, false
			}
		}
		return nil, true
	}

	acc := make([]core.Source, 0, len(sources))
	up := make([]*DummySource, 0, len(sources))
	for _, v := range sources {
		if p, ok := sf(v); ok {
			acc = append(acc, v)
		} else {
			up = append(up, &DummySource{
				internal: v,
				Name:     v.Name(),
				Policy:   p,
				Blocked:  true,
			})
		}
	}

	ss.protected.Put(acc...)
	if ss.underPolicy == nil {
		ss.underPolicy = make([]*DummySource, 0, len(up))
	}

	// Avoid adding duplicate values.
	dsf := func(src *DummySource) bool {
		for _, v := range ss.underPolicy {
			if v.Name == src.Name {
				return false
			}
		}
		return true
	}
	for _, v := range up {
		if dsf(v) {
			ss.underPolicy = append(ss.underPolicy, v)
		}
	}
}

// Del removes the policies from the protected storage and
// from the list of sources under policy.
func (ss *SourceStore) Del(sources ...core.Source) {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	ss.protected.Del(sources...)

	f := func(src *DummySource) bool {
		for _, v := range sources {
			if v.Name() == src.Name {
				return false
			}
		}
		return true
	}

	up := make([]*DummySource, 0, len(ss.underPolicy))
	for _, v := range ss.underPolicy {
		if f(v) {
			up = append(up, v)
		}
	}
	ss.underPolicy = up
}

// GetPoliciesSnapshot returns a copy of the current policies
// active in the store.
func (ss *SourceStore) GetPoliciesSnapshot() []*Policy {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	acc := make([]*Policy, len(ss.Policies))
	copy(acc, ss.Policies)
	return acc
}
