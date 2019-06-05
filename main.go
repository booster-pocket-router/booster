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

const gw = "10.12.44.10"

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

	tun, err := TUN()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully attached to TUN device: %s (MTU: %d)\n", tun.Name(), tun.MTU())

	ifconfig := Ifconfig{}
	if err := ifconfig.Up(tun.Name(), gw, gw).Run(); err != nil {
		panic(err)
	}

	p := make([]byte, tun.MTU())
	for {
		n, err := tun.Read(p)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("% x\n", p[:n])

		if !*echo {
			continue
		}
		n, err = tun.Write(p[:n])
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Bytes written: %d\n", n)
	}
}
