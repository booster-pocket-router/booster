## Notes

#### Forwarding packets from TUN interface to another interface
- Enabled packet forwarding on macOS
	`sudo sysctl -w net.inet6.ip6.forwarding=1`
	`sudo sysctl -w net.inet.ip.forwarding=1`
	`sysctl net | grep forward` # For checking the values
