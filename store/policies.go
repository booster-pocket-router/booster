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
)

const PolicyCodeBlock int = iota + 1

type basePolicy struct {
	// Reason explains why this policy exists.
	Reason string `json:"reason"`
	// Issuer tells where this policy comes from.
	Issuer string `json:"issuer"`
	// Code is the code of the policy, usefull when the policy
	// is delivered to another context.
	Code int `json:"code"`
	// Desc describes how the policy acts.
	Desc string `json:"description"`
}

// GenPolicy is a general purpose policy that allows
// to configure the behaviour of the Accept function
// setting its AcceptFunc field.
type GenPolicy struct {
	basePolicy

	Name string `json:"id"`
	// AcceptFunc is used as implementation
	// of Accept.
	AcceptFunc func(id, target string) bool `json:"-"`
}

// Accept implements Policy.
func (p *GenPolicy) Accept(id, target string) bool {
	return p.AcceptFunc(id, target)
}

// ID implements Policy.
func (p *GenPolicy) ID() string {
	return p.Name
}

// NewBlockPolicy returns a new Policy that blocks source `sourceID`.
func NewBlockPolicy(issuer, sourceID string) *BlockPolicy {
	return &BlockPolicy{
		basePolicy: basePolicy{
			Issuer: issuer,
			Code:   PolicyCodeBlock,
			Desc:   fmt.Sprintf("Source %v will no longer be used", sourceID),
		},
		SourceID: sourceID,
	}
}

// BlockPolicy is a specific policy used to block single sources.
type BlockPolicy struct {
	basePolicy
	// Source that should be always refuted.
	SourceID string `json:"-"`
}

// Accept implements Policy.
func (p *BlockPolicy) Accept(id, target string) bool {
	return id != p.SourceID
}

// ID implements Policy.
func (p *BlockPolicy) ID() string {
	return "block_" + p.SourceID
}
