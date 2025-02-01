[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub)
[![License](https://img.shields.io/github/license/merliot/hub)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)
[![Issues](https://img.shields.io/github/issues/merliot/hub)](https://github.com/merliot/hub/issues)

# Merliot
Private Device Hub

<a href="https://merliot.io/demo">
	<img src="pkg/device/docs/images/demo.svg" width="500px">
</a>

## Run from Docker

<pre>
$ sudo docker run -p 8000:8000 merliot/hub
</pre>

Browse to `http://localhost:8000` to view hub.

See other [install](https://merliot.io/doc/install) methods.

## Run on Cloud

Run a FREE hub instance on <a target="_blank" href="koyeb.com">Koyeb</a>.  Use the one-click button to get started:

<a href="https://app.koyeb.com/deploy?name=hub&type=docker&image=merliot%2Fhub&instance_type=free&regions=was&ports=8000;http;/&env[LOG_LEVEL]=INFO&env[PING_PERIOD]=2&env[BACKGROUND]=&env[DEVICES]=&env[USER]=&env[PASSWD]=&env[WIFI_SSIDS]=&env[WIFI_PASSPHRASES]=">
	<img src="https://www.koyeb.com/static/images/deploy/button.svg">
</a>

See other cloud [install](https://merliot.io/doc/install) methods.

## Run from Source

<pre>
$ git clone https://github.com/merliot/hub.git
$ cd hub
$ go run ./cmd
</pre>
