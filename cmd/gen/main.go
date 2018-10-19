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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/log"
)

const (
	// contains the initial list of sources used by booster.
	fileIfaces = "interfaces.json"
)

const (
	version = 1
)

// Version and BuildTime are filled in during build by the Makefile
var (
	Version   = "N/A"
	BuildTime = "N/A"
)

var verbose = flag.Bool("verbose", false, "Enable verbose mode")
var name = flag.String("name", "", "Collect only interfaces which name contains \"name\"")

func Dprintf(format string, a ...interface{}) {
	w := ioutil.Discard
	if *verbose {
		w = os.Stdout
	}
	fmt.Fprintf(w, format, a...)
}

func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func main() {
	flag.Parse()

	Printf("Version: %s, BuildTime: %s\n\n", Version, BuildTime)

	ifs := GetFilteredInterfaces(*name)
	if len(ifs) == 0 {
		Printf("No relevant interfaces found\n")
		return
	}

	Printf("\n\nCollected %d relevant interfaces:\n", len(ifs))
	for i, v := range ifs {
		Printf("%d: %+v\n", i, v)
	}
	Dprintf("\n")

	// Create sources file
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to open work directory: %v", err)
	}

	f := filepath.Join(dir, fileIfaces)
	Dprintf("Creating sources file %v\n", f)

	file, err := os.Create(f)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}
	defer file.Close()

	// Convert net.Interfaces into InterfaceSources

	// Encode collected interfaces & save them into the file
	var w io.Writer = file
	if *verbose {
		w = io.MultiWriter(w, os.Stdout)
	}

	var ifaces []*booster.IfaceSource
	for _, v := range ifs {
		i, err := mapIface(v)
		if err != nil {
			Printf("Unable to map net.Interface to booster.IfaceSource: %v. Skipping...", err)
			continue
		}
		ifaces = append(ifaces, i)
	}

	Dprintf("Encoding filtered interfaces into %v\n", f)
	if err = json.NewEncoder(w).Encode(struct {
		Version int                    `json:"version"`
		Body    []*booster.IfaceSource `json:"body"`
	}{
		Version: version,
		Body:    ifaces,
	}); err != nil {
		log.Fatalf("Unable to encode interfaces: %v", err)
	}
}

func GetFilteredInterfaces(s string) []net.Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		Printf("Unable to get interfaces: %v\n", err)
		return ifs
	}

	l := make([]net.Interface, 0, len(ifs))

	for i, v := range ifs {
		if i > 0 {
			Dprintf("\n")
		}
		Dprintf("Inspecting interface %+v\n", v)

		if len(v.HardwareAddr) == 0 {
			Dprintf("Empty hardware address. Skipping interface...\n")
			continue
		}

		if s != "" && !strings.Contains(v.Name, s) {
			Dprintf("Interface name does not satisfy name requirements: must contain \"%s\"\n", s)
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			// If the source does not contain an error
			Printf("Unable to get interface addresses: %v. Skipping interface...\n", err)
			continue
		}

		if len(addrs) == 0 {
			Dprintf("Empty unicast/multicast address list. Skipping interface...\n")
			continue
		}

		l = append(l, v)
	}

	return l
}

func mapIface(i net.Interface) (*booster.IfaceSource, error) {
	l, _ := i.Addrs() // has already been checked
	addrs := make([]string, 0, len(l))

	for _, v := range l {
		addr, err := resolveTCPAddr(v)
		if err != nil {
			// Ignore the issue and go on. We just one one address
			continue
		}

		addrs = append(addrs, addr.String())
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("Unable to find one resolvable IP address among: %+v", l)
	}

	return &booster.IfaceSource{
		Name:         i.Name,
		HardwareAddr: i.HardwareAddr.String(),
		Addrs:        addrs,
		MTU:          i.MTU,
		Flags:        i.Flags,
	}, nil
}

func resolveTCPAddr(addr net.Addr) (*net.TCPAddr, error) {
	ipAddr, _, err := net.ParseCIDR(addr.String())
	if err != nil {
		return nil, errors.New("Unable to parse CIDR: " + err.Error())
	}

	s := net.JoinHostPort(ipAddr.String(), "0")
	return net.ResolveTCPAddr("tcp", s)
}

