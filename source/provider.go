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

package source

import (
	"context"
	"fmt"

	"github.com/booster-proj/booster/core"
)

type Confidence int

const (
	Low Confidence = iota
	High
)

// Provider is a provider implementation which acts as a wrapper
// around many provider implementations.
type MergedProvider struct {
	// ControlInterface allows to make some final configurations
	// on an interface that has been found by the provider, before
	// it is hidden inside a core.Source.
	ControlInterface func(ifi *Interface)

	local *Local
}

// Provide returns the list of sources returned by each provider owned
// by merged. Currently only a local provider is queried.
func (p *MergedProvider) Provide(ctx context.Context) ([]core.Source, error) {
	if p.local == nil {
		p.local = new(Local)
	}

	interfaces, err := p.local.Provide(ctx, Low)
	if err != nil {
		return []core.Source{}, err
	}

	sources := make([]core.Source, 0, len(interfaces))
	for _, v := range interfaces {
		if f := p.ControlInterface; f != nil {
			f(v)
		}
		sources = append(sources, v)
	}
	return sources, nil
}

func (p *MergedProvider) Check(ctx context.Context, src core.Source, level Confidence) error {
	if ifi, ok := src.(*Interface); ok {
		return p.local.Check(ctx, ifi, level)
	}
	return fmt.Errorf("provider: unable to find suitable checks for source %s", src.Name())
}
