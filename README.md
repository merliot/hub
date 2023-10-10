# Merliot Hub

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)

![Gopher Thing](images/gopher_cloud.png)

Merliot Hub is a device hub.  It's written in [Go](go.dev) and [TinyGo](tinygo.org).

## Device Platforms

Merliot Hub supports devices created on these platforms:

- [Raspberry Pi 3/4](https://www.raspberrypi.com/)
- [Raspberry Pi Pico W](https://www.raspberrypi.com/documentation/microcontrollers/raspberry-pi-pico.html) (Coming soon!)
- [Arduino Nano Connect rp2040](https://docs.arduino.cc/hardware/nano-rp2040-connect)
- [Seeed Wio Terminal](https://www.seeedstudio.com/Wio-Terminal-p-4509.html)
- [Adafruit PyPortal](https://www.adafruit.com/product/4116)

## Quick Start

Deploy a Merliot Hub in your own [docker](https://www.docker.com/) environment:

```
git clone https://github.com/merliot/hub.git
cd hub
docker build -t hub -f Dockerfile-http .
docker run -p 80:8000 hub
```

Browse to [http://127.0.0.1](http://127.0.0.1) to view hub and create devices.

Or, one-click deploy a Merliot Hub on these cloud providers:

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=git&repository=github.com/merliot/hub&branch=main&name=hub&builder=dockerfile)

## Saving Changes

Merliot Hub saves device changes back to the hub repo.  To enable saving device changes, use your own copy of the repo:

1. [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo) this repo.
2. Build the docker image from the fork.

    ```
    git clone <fork path>/hub.git
    cd hub
    docker build -t hub -f Dockerfile-http .
    ```

4. Pass in to docker GIT_xxx enviroment vars:

    ```
    docker run -p 80:8000 -e GIT_AUTHOR=<author> -e GIT_KEY=<key> -e GIT_REMOTE=<remote> hub
    ```

(If using cloud provider, pass the GIT_xxx environment vars using the provider's sercrets to store the GIT_xxx values).
