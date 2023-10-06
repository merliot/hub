# Merliot Hub

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)

![Gopher Thing](images/gopher_cloud.png)

Merliot Hub is a device hub.  It's written in [Go](go.dev) and [TinyGo](tinygo.org).

## Quick Start

One-click deploy a Merliot Hub on these cloud providers:

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=git&repository=github.com/merliot/hub&branch=main&name=hub&builder=dockerfile)

Or, deploy a Merliot Hub in your own docker environment:

```
git clone https://github.com/merliot/hub.git
cd hub
docker build -tag hub -f Dockerfile-http .
docker run -p 80:8000 hub
```

Browse to [http://127.0.0.1](http://127.0.0.1) to view hub and create devices.

If you want an https version of the hub, use Dockerfile (not Dockerfile-http).

> [!NOTE]
> If you want to save changes, you'll need to fork the repo and run Docker from your fork.

## Support Platforms




