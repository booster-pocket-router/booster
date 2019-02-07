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
	"net"
	"time"
)

type HostResolver interface {
	// Returns all ip addresses associated with `host`
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	// Returns at least an host associated with `addr`
	LookupAddr(ctx context.Context, addr string) (hosts []string, err error)
}

var Resolver HostResolver = &net.Resolver{}

// Policy codes, different for each `Policy` created.
const (
	PolicyCodeBlock int = iota + 1
	PolicyCodeReserve
	PolicyCodeStick
	PolicyCodeAvoid
)

type basePolicy struct {
	Name string `json:"id"`

	// Reason explains why this policy exists.
	Reason string `json:"reason"`

	// Issuer tells where this policy comes from.
	Issuer string `json:"issuer"`

	// Code is the code of the policy, usefull when the policy
	// is delivered to another context.
	Code int `json:"code"`

	// Desc describes how the policy acts.
	Desc string `json:"description"`

	// Addrs is the list of address address that the
	// policy takes into consideration.
	Addrs []string `json:"addresses"`
}

func (p basePolicy) ID() string {
	return p.Name
}

// GenPolicy is a general purpose policy that allows
// to configure the behaviour of the Accept function
// setting its AcceptFunc field.
//
// Used mainly in tests.
type GenPolicy struct {
	basePolicy
	Name string `json:"name"`

	// AcceptFunc is used as implementation
	// of Accept.
	AcceptFunc func(id, address string) bool `json:"-"`
}

func (p *GenPolicy) ID() string {
	return p.Name
}

// Accept implements Policy.
func (p *GenPolicy) Accept(id, address string) bool {
	return p.AcceptFunc(id, address)
}

// BlockPolicy blocks `SourceID`.
type BlockPolicy struct {
	basePolicy
	// Source that should be always refuted.
	SourceID string `json:"-"`
}

func NewBlockPolicy(issuer, sourceID string) *BlockPolicy {
	return &BlockPolicy{
		basePolicy: basePolicy{
			Name:   "block_" + sourceID,
			Issuer: issuer,
			Code:   PolicyCodeBlock,
			Desc:   fmt.Sprintf("source %v will no longer be used", sourceID),
		},
		SourceID: sourceID,
	}
}

// Accept implements Policy.
func (p *BlockPolicy) Accept(id, address string) bool {
	return id != p.SourceID
}

// ReservedPolicy is a Policy implementation. It is used to reserve a source
// to be used only for connections to a defined address, and those connections
// will not be assigned to any other source.
type ReservedPolicy struct {
	basePolicy
	SourceID string `json:"reserved_source_id"`
}

func NewReservedPolicy(issuer, sourceID string, hosts ...string) *ReservedPolicy {
	addrs := []string{}
	for _, v := range hosts {
		address := TrimPort(v)
		addrs = append(addrs, LookupAddress(address)...)
	}
	return &ReservedPolicy{
		basePolicy: basePolicy{
			Name:   fmt.Sprintf("reserve_%s", sourceID),
			Issuer: issuer,
			Code:   PolicyCodeReserve,
			Desc:   fmt.Sprintf("source %v will only be used for connections to %v", sourceID, addrs),
			Addrs:  addrs,
		},
		SourceID: sourceID,
	}
}

// Accept implements Policy.
func (p *ReservedPolicy) Accept(id, address string) bool {
	isIn := false
	for _, v := range p.Addrs {
		if address == v {
			isIn = true
			break
		}
	}
	if isIn {
		return id == p.SourceID
	}

	return id != p.SourceID
}

// AvoidPolicy is a Policy implementation. It is used to avoid giving
// connection to `Address` to `SourceID`.
type AvoidPolicy struct {
	basePolicy
	SourceID string `json:"avoid_source_id"`
	Address  string `json:"address"`
}

func NewAvoidPolicy(issuer, sourceID, address string) *AvoidPolicy {
	address = TrimPort(address)
	return &AvoidPolicy{
		basePolicy: basePolicy{
			Name:   fmt.Sprintf("avoid_%s_for_%s", sourceID, address),
			Issuer: issuer,
			Code:   PolicyCodeAvoid,
			Desc:   fmt.Sprintf("source %v will not be used for connections to %s", sourceID, address),
			Addrs:  LookupAddress(address),
		},
		SourceID: sourceID,
		Address:  address,
	}
}

// Accept implements Policy.
func (p *AvoidPolicy) Accept(id, address string) bool {
	isIn := false
	for _, v := range p.Addrs {
		if address == v {
			isIn = true
			break
		}
	}
	if isIn {
		return id != p.SourceID
	}
	return true
}

// HistoryQueryFunc describes the function that is used to query the bind
// history of an entity. It is called passing the connection address in question,
// and it returns the source identifier that is associated to it and true,
// otherwise false if none is found.
type HistoryQueryFunc func(string) (string, bool)

// StickyPolicy is a Policy implementation. It is used to make connections to
// some address be always bound with the same source.
type StickyPolicy struct {
	basePolicy
	BindHistory HistoryQueryFunc `json:"-"`
}

func NewStickyPolicy(issuer string, f HistoryQueryFunc) *StickyPolicy {
	return &StickyPolicy{
		basePolicy: basePolicy{
			Name:   "stick",
			Issuer: issuer,
			Code:   PolicyCodeStick,
			Desc:   "once a source receives a connection to a address, the following connections to the same address will be assigned to the same source",
		},
		BindHistory: f,
	}
}

// Accept implements Policy.
func (p *StickyPolicy) Accept(id, address string) bool {
	if hid, ok := p.BindHistory(address); ok {
		return id == hid
	}

	return true
}

// TrimPort removes port information from `address`.
func TrimPort(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return host
	}
	return address
}

// LookupAddress finds the addresses associated with `address`. If it
// is not able to lookup, it just returns `address` wrapped into a list.
func LookupAddress(address string) []string {
	address = TrimPort(address)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	names, err := Resolver.LookupHost(ctx, address)
	if err != nil {
		return []string{address}
	}
	return names
}
