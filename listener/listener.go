/*
Copyright (C) 2018 KIM KeepInMind GmbH/srl

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

package listener

import (
	"context"
	"net"
	"strings"
	"fmt"
	"time"

	"github.com/booster-proj/booster/sources"
	"upspin.io/log"
)

type Storage interface {
	Put(... interface { ID() string })
	Del(... interface { ID() string })
}

type Listener struct {
	s Storage
}

func New(s Storage) *Listener {
	return &Listener{s}
}

var poolInterval = time.Second*5
var poolTimeout = time.Second*2

// Err is a Listener's critical error.
type Err struct {
	e error
}

func (e *Err) Error() string {
	return "critical: " + e.e.Error()
}

// filterErr either logs the error or it returns
// it, if it's critical.
func filterErr(err error) error {
	if _err, ok := err.(*Err); ok {
		return _err
	}

	log.Error.Printf("Listener error: %v", err)
	return nil
}

func (l *Listener) Run(ctx context.Context) error {
	// Pool first
	_ctx, cancel := context.WithTimeout(ctx, poolTimeout)
	defer cancel()

	if err := filterErr(l.Pool(_ctx)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			// Exit in case of context cancelation
			return ctx.Err()
		case <-time.After(poolInterval):
			_ctx, _cancel := context.WithTimeout(ctx, poolTimeout)
			if err := filterErr(l.Pool(_ctx)); err != nil {
				_cancel()
				return err
			}
			_cancel()

			// TODO: here the listener as an updated state after a
			// successfull pool.

			// Wait a fixe
			<-time.After(poolInterval)
		}
	}
}

func (l *Listener) Pool(ctx context.Context) error {
	return &Err{fmt.Errorf("Pool is not yet implemented")}
}

// TODO: Remove
func GetFilteredInterfaces(s string) []*sources.Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		log.Error.Printf("Unable to get interfaces: %v\n", err)
		return []*sources.Interface{}
	}

	l := make([]*sources.Interface, 0, len(ifs))

	for _, v := range ifs {
		log.Debug.Printf("Inspecting interface %+v\n", v)

		if len(v.HardwareAddr) == 0 {
			log.Debug.Printf("Empty hardware address. Skipping interface...")
			continue
		}

		if s != "" && !strings.Contains(v.Name, s) {
			log.Debug.Printf("Interface name does not satisfy name requirements: must contain \"%s\"", s)
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			// If the source does not contain an error
			log.Debug.Printf("Unable to get interface addresses: %v. Skipping interface...", err)
			continue
		}
		if len(addrs) == 0 {
			log.Debug.Printf("Empty unicast/multicast address list. Skipping interface...")
			continue
		}

		l = append(l, &sources.Interface{Interface: v})
	}

	return l
}
