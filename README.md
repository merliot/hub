[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub)
[![License](https://img.shields.io/github/license/merliot/hub)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)
[![Issues](https://img.shields.io/github/issues/merliot/hub)](https://github.com/merliot/hub/issues)

# Merliot
Private Device Hub

<img src="docs/images/hub-light.png" width="300px">

## Run from Docker

<pre>
$ sudo docker run -p 8000:8000 merliot/hub
</pre>

Browse to `http://localhost:8000` to view hub.

See other [install](https://merliot.io/doc/install) methods.

## Run from Source

<pre>
$ git clone https://github.com/merliot/hub.git
$ cd hub
$ go run ./cmd
</pre>
