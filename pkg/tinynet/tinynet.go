//go:build pyportal || nano_rp2040 || metro_m4_airlift || arduino_mkrwifi1010 || matrixportal_m4 || wioterminal

package tinynet

import (
	"errors"
	"net"
	"net/netip"
	"time"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var link netlink.Netlinker
var dev netdev.Netdever

func NetConnect(ssid, pass string) error {

	link, dev = probe.Probe()

	return link.NetConnect(&netlink.ConnectParams{
		Ssid:            ssid,
		Passphrase:      pass,
		WatchdogTimeout: 10 * time.Second,
	})
}

func GetHardwareAddr() (net.HardwareAddr, error) {
	if link == nil {
		return net.HardwareAddr{}, errors.New("Not available")
	}
	return link.GetHardwareAddr()
}

func Addr() (netip.Addr, error) {
	if dev == nil {
		return netip.Addr{}, errors.New("Not available")
	}
	return dev.Addr()
}
