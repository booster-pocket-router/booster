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
var redirect = flag.Bool("redirect", false, "Redirect here ALL network traffic")
var gw = flag.String("gw", "10.12.44.16", "IP address that will be assigned to the TUN device")

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

// Batch of:
// sudo route -n add 0/1 10.12.44.16
// sudo route -n add 128.0/1 10.12.44.16
// Tries to rollback in case of problems.
func (c Route) RedirectAll(gw string) error {
	net1 := "0/1"
	net2 := "128.0/1"
	rollback := func() {
		// We need to cleanup only if the second
		// command fails.
		c.Del(net1, gw).Run()
	}

	if err := c.Add(net1, gw).Run(); err != nil {
		return err
	}
	if err := c.Add(net2, gw).Run(); err != nil {
		rollback()
		return err
	}

	return nil
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

	iff, err := TUN()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully attached to TUN device: %s\n", iff.Name())
	fmt.Printf("MTU: %d\n", iff.MTU())

	ifconfig := Ifconfig{}
	route := Route{}

	// Bring the interface UP
	if err := ifconfig.Up(iff.Name(), *gw, *gw).Run(); err != nil {
		panic(err)
	}

	// Redirect all traffic here if required
	if *redirect {
		if err := route.RedirectAll(*gw); err != nil {
			panic(err)
		}
	}

	packet := make([]byte, iff.MTU())
	for {
		n, err := iff.Read(packet)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Packet received: % x\n", packet[:n])

		if !*echo {
			// Do not write packet back to interface if
			// not in "echo" mode
			continue
		}
		n, err = iff.Write(packet[:n])
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Bytes written: %d\n", n)
	}
}
