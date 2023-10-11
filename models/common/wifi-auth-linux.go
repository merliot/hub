//go:build !tinygo

package common

import (
	"os"
	"strconv"
)

func (c *Common) ParseWifiAuth() {
	if ssid, ok := os.LookupEnv("WIFI_SSID"); ok {
		if passphrase, ok := os.LookupEnv("WIFI_PASSPHRASE"); ok {
			c.WifiAuth[ssid] = passphrase
		}
	}
	for i := 0; i < 10; i++ {
		a := strconv.Itoa(i)
		if ssid, ok := os.LookupEnv("WIFI_SSID_" + a); ok {
			if passphrase, ok := os.LookupEnv("WIFI_PASSPHRASE_" + a); ok {
				c.WifiAuth[ssid] = passphrase
			}
		}
	}
}
