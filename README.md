# Merliot Hub

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)

Merliot Hub is a device hub.  It's written in [Go](go.dev) and [TinyGo](tinygo.org).

## Quick Start

The quick start is to run Merliot Hub in a docker container directly from the repo.  Clone this repo and build and run a docker image from the Dockerfile:

```
git clone https://github.com/merliot/hub.git
cd hub
```

We'll use the Dockerfile-http to build an image for http so we can run it locally.  (If you want secure https support, use Dockerfile).

```
docker build -tag hub -f Dockerfile-http .
docker run -p 80:8000 hub
```

Browse to [http://127.0.0.1](http://127.0.0.1) to view hub.  You can create devices and connect those devices to the hub, however changes will not be saved.  To save changes, you'll need to work from your own fork of this repo.

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=git&repository=github.com/merliot/hub&branch=main&name=hub)
