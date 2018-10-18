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

package main

import (
	"errors"
	"flag"
	"net"
	"strings"

	"github.com/booster-proj/log"
)

const (
	// contains the initial list of sources used by booster.
	fileSources = "booster.json"
)

// Version and BuildTime are filled in during build by the Makefile
var (
	Version   = "N/A"
	BuildTime = "N/A"
)

var verbose = flag.Bool("verbose", false, "Enable verbose mode")
var name = flag.String("name", "", "Collect only interfaces which name contains \"name\"")

func main() {
	flag.Parse()

	log.Info.Printf("Version: %s, BuildTime: %s\n\n", Version, BuildTime)
	if *verbose {
		log.Info.Printf("running in verbose mode")
		log.SetLevel(log.DebugLevel)
	}

	ifs := getFilteredInterfaces(*name)
	log.Println("")

	for i, v := range ifs {
		log.Info.Printf("%d: %+v", i, v)
	}
}

func getFilteredInterfaces(s string) []net.Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		log.Error.Printf("Unable to get interfaces: %v", err)
		return ifs
	}

	l := make([]net.Interface, 0, len(ifs))

	for _, v := range ifs {
		log.Debug.Printf("Inspecting interface %+v", v)

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
			log.Error.Printf("Unable to get interface addresses: %v. Skipping interface...", err)
			continue
		}

		if len(addrs) == 0 {
			log.Debug.Printf("Empty unicast/multicast address list. Skipping interface...")
			continue
		}

		l = append(l, v)
	}

	return l
}

func resolveTCPAddr(addr net.Addr) (*net.TCPAddr, error) {
	ipAddr, _, err := net.ParseCIDR(addr.String())
	if err != nil {
		return nil, errors.New("Unable to parse CIDR: " + err.Error())
	}

	s := net.JoinHostPort(ipAddr.String(), "0")
	return net.ResolveTCPAddr("tcp", s)
}
