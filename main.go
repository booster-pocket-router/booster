// +build darwin

package main

import (
	"log"
	"flag"
	"fmt"
	"os/exec"
	"os/user"
	"net"
	"errors"

	"github.com/songgao/water"
)

var echo = flag.Bool("echo", false, "Echo packets received back to TUN")

const gwIn = "10.12.44.10"
const gwOut = "10.12.44.20"

type Iff struct {
	*water.Interface
}

func (i *Iff) MTU() int {
	netIff, err := net.InterfaceByName(i.Name())
	if err != nil {
		return -1
	}
	return netIff.MTU
}

func TUN() (*Iff, error) {
	// Interface is not persistent
	wIff, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}

	return &Iff{
		Interface: wIff,
	}, nil
}

type Ifconfig struct {}

func (c Ifconfig) Name() string {
	return "ifconfig"
}

func (c Ifconfig) Up(name, dst, gw string) *exec.Cmd {
	return exec.Command(c.Name(), name, gw, dst, "up")
}

type Route struct {}

func (c Route) Name() string {
	return "route"
}

func (c Route) Add(dst, gw string) *exec.Cmd {
	return exec.Command(c.Name(), "-n", "add", dst, gw)
}

func (c Route) Del(dst, gw string) *exec.Cmd {
	return exec.Command(c.Name(), "-n", "del", dst, gw)
}

func CheckPermissions() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	if u.Username != "root" {
		return errors.New("This program requires root permissions")
	}
	return nil
}

func main() {
	flag.Parse()
	if err := CheckPermissions(); err != nil {
		panic(err)
	}

	tunIn, err := TUN()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully attached to TUN device: %s (MTU: %d)\n", tunIn.Name(), tunIn.MTU())

	tunOut, err := TUN()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully attached to TUN device: %s (MTU: %d)\n", tunOut.Name(), tunOut.MTU())

	ifconfig := Ifconfig{}
	if err := ifconfig.Up(tunIn.Name(), gwIn, gwIn).Run(); err != nil {
		panic(err)
	}
	if err := ifconfig.Up(tunOut.Name(), gwOut, gwOut).Run(); err != nil {
		panic(err)
	}

	s := max(tunIn.MTU(), tunOut.MTU())
	p := make([]byte, s)
	for {
		n, err := tunIn.Read(p)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Packet received: % x\n", p[:n])

		if !*echo {
			continue
		}
		n, err = tunOut.Write(p[:n])
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Bytes written: %d\n", n)
	}
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
