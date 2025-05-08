[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub/pkg/device)
[![License](https://img.shields.io/github/license/merliot/hub)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)
[![Issues](https://img.shields.io/github/issues/merliot/hub)](https://github.com/merliot/hub/issues)
[![codecov](https://codecov.io/gh/merliot/hub/graph/badge.svg?token=N0ATO7YP4U)](https://codecov.io/gh/merliot/hub)

# MERLIOT
PRIVATE DEVICE HUB

## Introduction

</p><a target="_blank" href="https://merliot.io">Merliot</a> Hub is a private device hub, a web app, and an MCP server.</p>

</p>There is a lot to unpack there.  Let’s start with the private device hub
part.</p>

### Private Hub

</p>Consumer-grade Smart devices such as Smart security cameras, Smart lights,
and Smart thermostats have one thing in common: they dial home to the device
manufacturer's central hub.  For each manufacture, you'll install a phone app
which connects to the same central hub.  Your data is unencrypted on the
manufacture's side of the hub.  Your data is not private.  The device
manufacturer can analyze, track, store, share, and sell your data.  Your data
plus your profile associated with the device purchase and app signup make _you_
the product.</p>

</p>Merliot Hub is private.  It’s private by switching from a central-hub
architecture to a distributed-hub architecture.  There is no central Merliot
Hub.  Each user of Merliot Hub installs their own hub.  A hub can be installed
on a local resource such as a laptop or Raspberry Pi.  Alternatively, it can be
installed on the cloud (for free in some cases).  Each user’s hub is
independent from others’.  There is no central hub.  No central point to
compromise, tap, exploit, or profit from.</p>

</p>The trade off for privacy is convenience.  Not only do you have to <a
target="_blank" href="https://www.merliot.io/doc/quick-start">install</a> your
own Merliot Hub, you have to build your own devices.  Merliot Hub devices are
built from hobbyist-level components that are readily available.  There are
camera, gps, relay, timer and other devices you can build for the hub.
Assembling the devices requires maker-level skills.  There is no software to
write, unless you want to create a new device model.  (Fun fact: <a
target="_blank" href="https://www.merliot.io/blog/2025-5-4-third-blog">LLMs</a>
can also write new device models).</p>

<div style="text-align: center;">
  <img src="devices/relays/images/nano-rp2040-relays.png">
</div>

### Web App

</p>Merliot Hub is a web app.  There is no phone app.  But, you can use any
browser on any device, including your phone, to access your hub, locally or
over the internet.  Try the demo to get a feel for the UI.</p>

<a href="https://merliot.io/demo">
	<img src="pkg/device/docs/images/demo.svg" width="500px">
</a>

### MCP Server

</p>Merliot Hub is a Model Context Protocol (<a target="_blank" src="doc/mcp-server">MCP</a>) server.  The MCP server lets
you plug your Merliot Hub into LLM hosts such as Claude or Cursor.  From there,
you can chat with the hub using natural language:</p>

<pre>
"List all of the devices in a tree view"
"Add a new gps device"
"Turn on all the relays"
"Show the instructions on how to deploy a qrcode device"
</pre>

</p>You can download a MCP server for a Merliot Hub by clicking the ‘tools’ icon.</p>

<img src="pkg/device/images/download-mcp-server.png">

## QUICK INSTALL

### &#x2B50; Install with Docker

<pre>
$ sudo docker run -p 8000:8000 merliot/hub
</pre>

Browse to `http://localhost:8000` to view hub.

See other [install](https://merliot.io/doc/install) methods.

### &#x2B50; Install on Cloud

Run a FREE hub instance on <a target="_blank" href="koyeb.com">Koyeb</a>.  Use the one-click button to get started:

<a href="https://app.koyeb.com/deploy?name=hub&type=docker&image=merliot%2Fhub&instance_type=free&regions=was&ports=8000;http;/&env[LOG_LEVEL]=INFO&env[PING_PERIOD]=2&env[BACKGROUND]=&env[DEVICES]=&env[USER]=&env[PASSWD]=&env[WIFI_SSIDS]=&env[WIFI_PASSPHRASES]=&env[AUTO_SAVE]=false">
	<img src="https://www.koyeb.com/static/images/deploy/button.svg">
</a>

See other cloud [install](https://merliot.io/doc/install) methods.

### &#x2B50; Run from Source

<pre>
$ git clone https://github.com/merliot/hub.git
$ cd hub
$ go run ./cmd
</pre>

Browse to http://localhost:8000 to view hub.

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
