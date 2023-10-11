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

```
git clone https://github.com/merliot/hub.git
cd hub
go run cmd/hub/main.go
```

Browse to [http://127.0.0.1:8000](http://127.0.0.1:8000) to view hub and deploy devices.

> [!NOTE]
> Deploying to TinyGo devices will not work.  Current limitation will be resolved in future TInyGo release.  To deploy on TinyGo devices, use the Docker or cloud methods below.

## Quick Start Docker

Deploy a Merliot Hub in your own [docker](https://www.docker.com/) environment:

```
git clone https://github.com/merliot/hub.git
cd hub
docker build -t hub -f Dockerfile .
docker run -p 80:8000 hub
```

Browse to [http://127.0.0.1](http://127.0.0.1) to view hub and deploy devices.

## Quick Start Cloud

One-click deploy a Merliot Hub on one of these cloud providers:

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=git&repository=github.com/merliot/hub&branch=main&name=hub&builder=dockerfile&env[SCHEME]=https)

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

(If using cloud provider, pass the GIT_xxx environment vars using the provider's secrets to store the GIT_xxx values).

## Environment Variables

The docker container looks for these environment vars on build and startup to configure the hub:

#### SCHEME
Scheme used for hub, either 'http' or 'https'.  Default is 'http'.

#### PORT
Port the hub listens on, default is 8000.

#### GIT_AUTHOR, GIT_REMOTE, GIT_KEY
Required if saving device changes.

#### BACKUP_HUB
Run as a backup hub.  A backup hub cannot make changes or deploy devices, but does provide an alternate address for viewing the hub devices.

#### BACKUP_HUB_URL
Set the backup hub URL.

#### USER, PASSWD
Set user and password for HTTP Basic Authentication on the hub.

#### WIFI_SSID, WIFI_PASSPHRASE
Set Wifi SSID and passphrase for devices built with TinyGo.  If mulitple SSID/passphrases are needed, use env vars WIFI_SSID_x and WIFI_SSID_PASSPHRASE_x, where x is 0-9.

## Building New Devices

New devices can be built from scratch or by extending existing devices.  The new device is given a unique model name.
