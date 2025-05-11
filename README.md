[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub/pkg/device)
[![License](https://img.shields.io/github/license/merliot/hub)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)
[![Issues](https://img.shields.io/github/issues/merliot/hub)](https://github.com/merliot/hub/issues)
[![codecov](https://codecov.io/gh/merliot/hub/graph/badge.svg?token=N0ATO7YP4U)](https://codecov.io/gh/merliot/hub)

# MERLIOT
AI DEVICE HUB

## INTRODUCTION

<a target="_blank" href="https://merliot.io">Merliot</a> Hub is an
AI-integrated device hub.

What does that mean?  It means you can control and interact with your physical
devices, your security camera, your thermometer, seamlessly using natural
language from an LLM host such as Claude Desktop or Cursor.  The hub is a
gateway between AI and the physical world.

What could go wrong?

<diagram>

Which devices?  Not the retail Smart devices, sorry.  You build a device from
hobby-grade components which are readily available, like Raspberry Pis,
buttons, relays, and sensors.  You need maker-level skills to build the
devices.  The hub includes a parts list and instructions to build each device.
There is no software to write; the hub includes the download-able device
firmware.

## FEATURES

- ### Privacy
    - Merliot Hub uses a distributed architecture rather than a centralized
architecture, eliminating third-party access to your devices' data. You
install and maintain your own hub.  No one else has access.  You data is
private.  <a target="_blank" href="">Read more</a>.

- ### Web App
    - Merliot hub is a web app.  There is no phone app.  But, you can use any
browser on any device, including your phone, to access your hub, locally
or over the internet.  Try the demo to get a feel for the UI.

<a target="_blank" href="https://merliot.io/demo">
	<img src="pkg/device/docs/images/demo.svg" width="500px">
</a>

- ### AI-Integration
	- Merliot Hub is a Model Context Protocol (<a target="_blank"
href="https://www.merliot.io/doc/mcp-server">MCP</a>) server.  The MCP server
lets you plug your Merliot Hub into LLM hosts such as Claude or Cursor.  From
there, you can chat with the hub using natural language:

		<pre>
		"List all of the devices in a tree view"
		"Add a new gps device"
		"Turn on all the relays"
		"Show the instructions on how to deploy a qrcode device"
		</pre>

- ### Docker










That is a lot to unpack.  Let’s start with the private device hub part.

### Private Hub

Consider consumer-grade Smart devices such as Smart security cameras, Smart
lights, and Smart thermostats.  These are not private.  They have one thing in
common: they dial home to the device manufacturer's central hub.  For each
manufacture, you'll install a phone app which connects to the same central hub.
Your data is unencrypted on the manufacture's side of the hub.  Your data is
not private.  The device manufacturer can analyze, track, store, share, and
sell your data.  Your data plus your profile associated with the device
purchase and app signup make _you_ the product.

Merliot Hub is private alternative.  It’s private by switching from a
central-hub architecture to a distributed-hub architecture.  There is no
central Merliot Hub.  Each user of Merliot Hub installs their own hub.  A hub
can be installed on a local resource such as a laptop or Raspberry Pi.
Alternatively, it can be installed on the cloud (for free in some cases).  Each
user’s hub is independent from others’.  There is no central hub.  No central
point to compromise, tap, exploit, or profit from.

The trade off for privacy is convenience.  You must <a target="_blank"
href="https://www.merliot.io/doc/quick-start">install</a> and maintain your own
Merliot Hub and you must to build your own devices.  Merliot Hub devices are
built from hobbyist-level components that are readily available.  There are
camera, gps, relay, timer and other devices you can build for the hub.
Assembling the devices requires maker-level skills.  There is no software to
write, unless you want to create a new device model.  (Fun fact: <a
target="_blank" href="https://www.merliot.io/blog/2025-5-4-third-blog">LLMs</a>
can write device models).

<div style="text-align: center;">
  <img src="devices/relays/images/nano-rp2040-relays.png">
</div>

### Web App

Merliot Hub is a web app.  There is no phone app.  But, you can use any
browser on any device, including your phone, to access your hub, locally or
over the internet.  Try the demo to get a feel for the UI.

<a target="_blank" href="https://merliot.io/demo">
	<img src="pkg/device/docs/images/demo.svg" width="500px">
</a>

### MCP Server

Merliot Hub is a Model Context Protocol (<a target="_blank"
href="https://www.merliot.io/doc/mcp-server">MCP</a>) server.  The MCP server
lets you plug your Merliot Hub into LLM hosts such as Claude or Cursor.  From
there, you can chat with the hub using natural language:

<pre>
"List all of the devices in a tree view"
"Add a new gps device"
"Turn on all the relays"
"Show the instructions on how to deploy a qrcode device"
</pre>

## SUPPORTED DEVICES TARGETS

Merliot Hub devices are build from one or more target platforms:

- <a href="https://www.raspberrypi.com/">Raspberry Pi (models 3, 4, 5)</a>
- <a href="https://store.arduino.cc/products/arduino-nano-rp2040-connect">Arduino Nano rp2040 Connect</a>
- <a href="https://www.adafruit.com/product/4116">Adafruit PyPortal</a>
- <a href="https://koyeb.com">Koyeb (cloud)</a>
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

Run a FREE hub instance on <a target="_blank" href="koyeb.com">Koyeb</a>.  Use this one-click button to get started:

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

I'd like to see others build cool devices to share.

There are always TODOs in the code that need attention.

## TESTING

<pre>
$ go test ./test/...
</pre>

## LICENSE

BSD 3-Clause License

## CONTACT

Email: <a target="_blank" href="mailto:contact@merliot.io">contact@merliot.io</a>

X: <a target="_blank" href="https://x.com/merliotio">@merliotio</a>

Slack: <a target="_blank" href="https://merliotcommunity.slack.com/messages/C06Q6QV6YSJ">#merliot</a>

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
