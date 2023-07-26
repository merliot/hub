package id

import (
	"net"
	"strings"
)

// Make up an id using the MAC address of the first non-lo interface
func MAC() string {
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Name != "lo" {
				addr := iface.HardwareAddr.String()
				return strings.Replace(addr, ":", "_", -1)
			}
		}
	}
	return "unknown"
}
