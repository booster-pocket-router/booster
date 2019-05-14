package main

import (
	"log"
	"flag"
	"fmt"
	"os/exec"
	"net"

	"github.com/songgao/water"
)

var echo = flag.Bool("echo", false, "Echo packets received back to TUN")
const ipv4A = "10.12.44.16"

type Iff struct {
	wIff *water.Interface
	netIff *net.Interface
	SrcAddr string
	DstAddr string
}

func (i *Iff) Up(src, dest string) error {
	return exec.Command("ifconfig", i.Name(), src, dest, "up").Run()
}

func (i *Iff) Read(p []byte) (int, error) {
	return i.wIff.Read(p)
}

func (i *Iff) Write(p []byte) (int, error) {
	return i.wIff.Write(p)
}

func (i *Iff) Name() string {
	return i.wIff.Name()
}

func (i *Iff) MTU() int {
	return i.netIff.MTU
}

func TunDev() (*Iff, error) {
	// Interface is not persistent
	wIff, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return nil, err
	}

	netIff, err := net.InterfaceByName(wIff.Name())
	if err != nil {
		return nil, err
	}

	return &Iff{
		wIff: wIff,
		netIff: netIff,
	}, nil
}

func main() {
	flag.Parse()

	iff, err := TunDev()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully attached to TUN device: %s\n", iff.Name())
	fmt.Printf("MTU: %d\n", iff.MTU())

	// Bring the interface UP
	if err := iff.Up(ipv4A, ipv4A); err != nil {
		panic(err)
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
