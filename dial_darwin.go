// +build darwin

package booster

import (
	"context"
	"errors"
	"net"
	"syscall"

	"upspin.io/log"
	"golang.org/x/sys/unix"
)

func (i Interface) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Find a suitable socket address from the interface
	var addr unix.Sockaddr

	addrs, err := i.Addrs()
	if err != nil {
		return nil, errors.New("Unable to retrieve interface addresses from interface " + i.Name + ": " + err.Error())
	}

	for _, v := range addrs {
		ip, _, err := net.ParseCIDR(v.String())
		if err != nil {
			return nil, errors.New("Unable to parse CIDR from interface " + i.Name + ": " + err.Error())
		}

		if ip4 := ip.To4(); ip4 != nil {
			// IPv4
			var buf [4]byte
			copy(buf[:], ip4[:4])
			addr = &unix.SockaddrInet4{
				Port: 0,
				Addr: buf,
			}

			log.Debug.Printf("Socket address for interface %v: %+v", i.Name, addr)

			break
		}
		// TODO(jecoz): Support ipv6
	}

	if addr == nil {
		return nil, errors.New("Unable to create a valid socket address from interface " + i.Name)
	}

	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if err := unix.Bind(int(fd), addr); err != nil {
					log.Error.Printf("Unable to bind to interface %v: %v", i.Name, err)
				}
			})
		},
	}

	return d.DialContext(ctx, network, address)
}
