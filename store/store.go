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
	// ID is used to identify later a policy.
	ID string
	// Func is the function used to check wether this policy
	// is applied to item with name == name or not. Returns
	// true if the input should be blocked/not accepted.
	Func func(name string) bool
	// Reason explains why this policy is applied, or who is
	// the issues of this policy. In other words, it explains
	// why this policy exists.
	Reason string
	// Code is the code of the policy, usefull when the policy
	// is delivered to another context.
	Code int
}

// A SourceStore is able to keep sources under a set of
// policies, or rules. When it is asked to store a value,
// it performs the policy checks on it, and eventually the
// request is forwarded to the protected store.
type SourceStore struct {
	protected Store

	Policies    []*Policy
	underPolicy []*DummySource
}

// A DummySource is a source which stores only the information
// of it's parent source at copy time, but it is no longer able
// to produce any internet conneciton. It should be used to show
// snapshots of the current storage to other componets of the
// program that should not be able to break or work with the
// original and active source.
type DummySource struct {
	internal core.Source            `json:"-"`
	Name     string                 `json:"name"`
	Policy   *Policy                `json:"policy"`
	Blocked  bool                   `json:"blocked"`
	Metrics  map[string]interface{} `json:"metrics"`
}

func New(store Store) *SourceStore {
	return &SourceStore{
		protected:   store,
		Policies:    []*Policy{},
		underPolicy: []*DummySource{},
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

// GetSourcesSnapshot returns a copy of the current sources that the store
// is handling. The sources returned are not capable of providing any
// internet connection, but are filled with the policies applied on
// them and the metrics collected.
func (ss *SourceStore) GetSourcesSnapshot() []*DummySource {
	acc := make([]*DummySource, 0, ss.protected.Len()+len(ss.underPolicy))

	ss.protected.Do(func(src core.Source) {
		ds := &DummySource{
			Name:    src.Name(),
			Blocked: false,
		}
		if metrics, ok := src.Value("metrics").(map[string]interface{}); ok {
			ds.Metrics = metrics
		}
		acc = append(acc, ds)
	})

	for _, v := range ss.underPolicy {
		ds := &DummySource{
			Name:    v.Name,
			Blocked: v.Blocked,
			Policy:  v.Policy,
		}
		if metrics, ok := v.internal.Value("metrics").(map[string]interface{}); ok {
			ds.Metrics = metrics
		}
		acc = append(acc, ds)
	}

	return acc
}

// Add policy stores the policy and applies it also to the sources
// stored in the protected storage, removing them from it if
// required.
func (ss *SourceStore) AddPolicy(p *Policy) {
	if ss.Policies == nil {
		ss.Policies = make([]*Policy, 0, 1)
	}
	ss.Policies = append(ss.Policies, p)

	// Now apply the new policy to the items that
	// are already in the storage.
	acc := make([]core.Source, 0, ss.protected.Len())
	ss.protected.Do(func(src core.Source) {
		if !p.Func(src.Name()) {
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
			Blocked:  true,
			Policy:   p,
		})
	}
}

// DelPolicy removes the policy with identifier id from the storage.
// It then loops through each source under policy, and frees it if
// the policy is the removed one, putting the source again in the
// protected storage.
// Note that only the first instance of policy with identifier id is
// removed.
func (ss *SourceStore) DelPolicy(id string) {
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
			ss.Put(v.internal)
		} else {
			acc = append(acc, v)
		}
	}
	ss.underPolicy = acc
}

func (rs *SourceStore) put(sources ...core.Source) {
	rs.protected.Put(sources...)
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
