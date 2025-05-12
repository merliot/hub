[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub/pkg/device)
[![License](https://img.shields.io/github/license/merliot/hub)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)
[![Issues](https://img.shields.io/github/issues/merliot/hub)](https://github.com/merliot/hub/issues)
[![codecov](https://codecov.io/gh/merliot/hub/graph/badge.svg?token=N0ATO7YP4U)](https://codecov.io/gh/merliot/hub)

# MERLIOT DEVICE HUB

<a href="https://merliot.io">Merliot</a> Hub is an
AI-integrated device hub.

What does that mean?  It means you can control and interact with your physical
devices, your security camera, your thermometer, seamlessly using natural
language from an LLM host such as [Claude Desktop](https://claude.ai/download)
or [Cursor](https://cursor.com).  The hub is a gateway between AI and the
physical world.

What could go wrong?

<a href="pkg/device/docs/images/mcp-server.gif">
    <img src="pkg/device/docs/images/mcp-server.gif">
</a>

### Devices

Which devices?  Not any of the retail Smart devices, sorry.  You build a device
from hobby-grade components which are readily available, like Raspberry Pis,
Arduinos, buttons, relays, and sensors.  You need maker-level skills to build
the devices.  The hub includes a parts list and instructions for building each
device.  There is no software to write; the hub includes the downloadable
device firmware.

<img src="devices/camera/images/rpi-camera.png" width="200">

## FEATURES

- ### Privacy

    - Merliot Hub uses a distributed architecture rather than a centralized
      architecture, eliminating third-party access to your devices' data. You
      install and maintain your own hub and devices.  No one else has access.  Your
      data is private.  Your data can't be sold, shared, stored, analyzed, or
      surveilled by third parties.  [Read more](https://www.merliot.io/doc/privacy).

- ### Web App

    - Merliot hub is a web app.  There is no phone app.  You can use any
      web browser on any device, including your phone, to access your hub,
      locally or over the internet.  Try the [demo](https://merliot.io/demo)
      to get a feel for the UI.

<a href="https://merliot.io/demo">
    <img src="pkg/device/docs/images/demo.svg" width="500px">
</a>

- ### AI-Integration

	- Merliot Hub is a Model Context Protocol ([MCP](https://www.merliot.io/doc/mcp-server))
      server.  The MCP server lets you plug your Merliot Hub into LLM hosts such as
      Claude or Cursor.  From there, you can chat with the hub using natural language:

		<pre>
		"List all of the devices in a tree view"
		"Add a new gps device"
		"Turn on all the relays"
		"Show the instructions on how to deploy a qrcode device"
		</pre>

- ### Cloud-Ready

    - Merliot Hub is packaged as a Docker image so you can run your hub
      anywhere you can run a Docker image, locally on your own laptop or server
      using Docker Desktop, or on the cloud.  See
      [install](https://merliot.io/doc/install) guide for more info.  The docker
      image requires a minimal VM: 0.1vCPU, 256MB RAM, 256MB disk.  Koyeb offers a
      [FREE](#install-on-cloud) VM suitable for running a hub in the cloud.

## SUPPORTED DEVICES TARGETS

Merliot Hub devices are built from one or more target platforms:

- [Raspberry Pi (models 3, 4, 5, and Zero 2W)](https://www.raspberrypi.com/)
- [Arduino Nano rp2040 Connect](https://store.arduino.cc/products/arduino-nano-rp2040-connect)
- [Adafruit PyPortal](https://www.adafruit.com/product/4116)
- [Koyeb (cloud)](https://koyeb.com)
- Linux x86-64

## QUICK START

See the official [Quick Start](https://merliot.io/doc/quick-start) and
[Install](https://merliot.io/doc/install) guides for more info.

### Install with Docker

<pre>
$ sudo docker run -p 8000:8000 merliot/hub
</pre>

Browse to `http://localhost:8000` to view hub.

### Install on Cloud

Run a FREE hub instance on [Koyeb](https://koyeb.com).  Use this one-click button to get started:

<a href="https://app.koyeb.com/deploy?name=hub&type=docker&image=merliot%2Fhub&instance_type=free&regions=was&ports=8000;http;/&env[LOG_LEVEL]=INFO&env[PING_PERIOD]=2&env[BACKGROUND]=&env[DEVICES]=&env[USER]=&env[PASSWD]=&env[WIFI_SSIDS]=&env[WIFI_PASSPHRASES]=&env[AUTO_SAVE]=false">
	<img src="https://www.koyeb.com/static/images/deploy/button.svg">
</a>

### Run from Source

<pre>
$ git clone https://github.com/merliot/hub.git
$ cd hub
$ go run ./cmd
</pre>

Browse to `http://localhost:8000` to view hub.

## CONTRIBUTING

PRs/Issues welcomed.

I'd like to see others build cool devices to share and to add to the project.

## TESTING

<pre>
$ go test ./test/...
</pre>

## LICENSE

BSD 3-Clause License

## CONTACT

Email: <a href="mailto:contact@merliot.io">contact@merliot.io</a>

X: [@merliotio](https://x.com/merliotio)

Slack: [#merliot](https://merliotcommunity.slack.com/messages/C06Q6QV6YSJ)

## CREDITS

Merliot is written in
	<a class="no-underline" href="https://go.dev/">Go</a>,
	<a class="no-underline" href="https://tinygo.org/">TinyGo</a>, and
	<a class="no-underline" href="https://htmx.org/">htmx.</a>
	Thank you to those who built and maintain these fine tools.

<div style="display: flex;">
	<a href="https://go.dev"><img src="pkg/device/docs/images/go-logo.png"></a>
	<a href="https://tinygo.org"><img src="pkg/device/docs/images/tinygo-logo.png"></a>
	<a href="https://htmx.org"><img src="pkg/device/docs/images/htmx-logo.png"></a>
</div>
