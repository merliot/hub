<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/images/merliot-hub-dark.png">
    <img src="docs/images/merliot-hub-light.png" width="70%">
  </picture>
</div>

<br><br>

<div align="center">

[![Go Reference](https://pkg.go.dev/badge/pkg.dev.go/github.com/merliot/hub.svg)](https://pkg.go.dev/github.com/merliot/hub)
[![License](https://img.shields.io/github/license/merliot/hub)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/merliot/hub)](https://goreportcard.com/report/github.com/merliot/hub)
[![Issues](https://img.shields.io/github/issues/merliot/hub)](https://github.com/merliot/hub/issues)

</div>

<h1 align="center">Your Private, Decentralized IoT Device Hub</h1>
<div align="center">
Securely access your devices from anywhere on the Internet · No app required
</div>
</br>


<p align="center">
    <a href="https://github.com/merliot/hub/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+">Report Bug</a>
    ·
    <a href="https://github.com/merliot/hub/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+">Request Feature</a>
    ·
  <a href="https://join.slack.com/t/merliotcommunity/shared_invite/zt-2f5f2t02q-jEmblYUmsQOxczvf6oJl8A">Join Our Slack</a>
    ·
    <a href="https://twitter.com/merliotio">Twitter</a>
</p>

<div align="center"><img src="docs/images/phone.png"></div>

## Features

- **Privacy first:** no user data collected or stored; no cookies; no tracking; no ads
- **Decentralized:** your hub is independent from your neighbor's; there is no central control over any hub
- **Secure:** TLS-secure communication from device and hub, and from hub to browser
- **Device targets:** target popular SBCs and microcontrollers such as Raspberry Pi and Arduino
- **No app to install:** runs as a responsive, single page web application; all you need is a browser
- **Containerized:** runs in a docker container; no dependencies to install
- **Small footprint:** minimum hardware specification is 0.1 vCPU, 256MB of RAM and 10GB of disk space.
- **100% Open Source**: written in [Go](https://go.dev) and [TinyGo](https://tinygo.org) (and some html/css/javascript for the UI)

## Quick Start

### Free on Koyeb

Click the button to install Merliot Hub on the Koyeb cloud (a Koyeb account is required).

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=docker&image=merliot/hub&name=hub&env[WS_SCHEME]=wss://)

Review the settings and click Apply.  It takes a few minutes for the hub to start.  Browse to `https://APP.koyeb.app/` to view hub and deploy devices.

### Docker

```
docker pull merliot/hub
docker run -p 80:8000 merliot/hub
```

Browse to `http://<host>` to view hub and deploy devices, where `<host>` is your IP address or hostname of your computer.


* [Install](#install)
  * [Local Install](#local-install)
  * [Cloud Install](#cloud-install)
  * [Local *and* Cloud Install](#local-and-cloud-install)
  * [Install from Source](#install-from-source)
* [Devices](#devices)
  * [Saving Devices](#saving-devices)
  * [Supported Targets](#supported-targets)
  * [Example Devices](#example-devices)
  * [Making a New Device](#making-a-new-device)
* [Environment Variables](#environment-variables)

## Install

Install Merliot Hub locally on your computer, on the cloud, or both, using our Docker image, without having to install all the dependencies.  (If you don't have [Docker](https://www.docker.com/), you can install the hub from [source](#install-from-source)).

### Local Install

Install Merliot Hub on a computer on your local network.  The devices will dial into the hub on your local network.  You access the hub at it's local IP address.

![](docs/images/local-install.png)

> [!NOTE]
> Prerequisite: Installed [Docker](https://docs.docker.com/get-docker/) environment.
  
```
docker pull ghcr.io/merliot/hub
docker run -p 8000:8000 merliot/hub
```

Browse to `http://<host>:8000` to view hub and deploy devices, where `<host>` is your IP address or hostname of your computer.

You can pass in [environment variables](#environment-variables).  For example, to set the Wifi [SSID/Passphrase](#wifi_ssids-wifi_passphrases) to be programmed into the devices:

```
docker run -e WIFI_SSIDS="My SSID" -e WIFI_PASSPHRASES="mypassphrase" -p 8000:8000 merliot/hub
```

Or to protect your hub with a [user/password](#user-passwd):

```
docker run -e USER="xxx" -e PASSWD="yyy" -p 8000:8000 merliot/hub
```

### Cloud Install

You can install Merliot Hub on the Internet using a cloud providers such as [Koyeb](https://www.koyeb.com), [Digital Ocean](https://www.digitalocean.com/), and [GCP](https://cloud.google.com) (Google Cloud Platform), to name a few.  The docker image path is:

```
docker pull ghcr.io/merliot/hub
```

![](docs/images/cloud-install.png)

#### Environment Variables

`PORT=8000`.  The hub listens on port :8000.

`WS_SCHEME=wss://`.  This uses the secure websocket scheme to connect to the hub.

(See additional [environment variables](#environment-variables)).

#### Install on Koyeb for Free

Click the button to install Merliot Hub on Koyeb, for Free!  A Koyeb account is required.

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=docker&image=merliot/hub&name=hub&env[WS_SCHEME]=wss://)

All cloud providers require an account, there's no getting around that.  Some have free-tiers or introductory credits to get started.  [Koyeb](https://www.koyeb.com) offers a free virtual machine with more than enough resources to run a hub.

Review the settings for the virtual machine (VM) and click Apply.  It takes a few minutes for the VM to start.  Your new hub will have an Internet URL in the format:

`https://hub-ACCOUNT.koyeb.app/`

Where ACCOUNT is your Koyeb account name.

> [!TIP]
> If you own a domain name, you can map it to the hub URL.

### Local *and* Cloud Install

Install Merliot Hub on a local computer *and* on the cloud, and the devices will dial into both.

On one hub, call it primary, set [`BACKUP`](#backup) environment to the URL of a backup hub.  Do the oposite, setting [`BACKUP`](#backup) on backup to point to primary's URL.  This way, regardless of which hub a device is created on, the device will dial into both hubs.

![](docs/images/local-and-cloud-install.png)

### Install from Source

> [!NOTE]
> Prerequisites:
> * [Go](https://go.dev/doc/install) version 1.22 or higher
> * [TinyGo](https://tinygo.org/getting-started/install/) version 0.31.1 or higher.

```
git clone https://github.com/merliot/hub.git
cd hub
go run ./cmd
```

Browse to `http://<host>` to view hub and deploy devices, where `<host>` is your IP address or hostname of your computer.

You can pass in [environment variables](#environment-variables).  For example, to set the [user/passwd](#user-passwd):

```
USER=foo PASSWD=bar go run ./cmd`
```

## Devices

A device is a gadget you build.  The picture-equation for a device is:

![device](images/device.png)

A device comprises a platform, some I/O, and the software (firmware) that runs on the device.  In this picture, the Raspberry Pi is the platform, the I/O is the relay and flow meter.  The device control code is written in Go; the device view code is written in HTML/JS/CSS.

The device dials into the hub so you can monitor and control the device from the hub.  Multiple devices, of different types, can dial into the hub.

The device is also a local web server, so you can browse directly to the device's address, skipping the hub.

### Saving Devices

### Supported Targets

Merliot Hub supports devices created on these platforms:

- [Raspberry Pi 3/4/5/Zero W/Zero 2 W](https://www.raspberrypi.com/)
- [Arduino Nano Connect rp2040](https://docs.arduino.cc/hardware/nano-rp2040-connect)
- [Seeed Wio Terminal](https://www.seeedstudio.com/Wio-Terminal-p-4509.html)
- [Adafruit PyPortal](https://www.adafruit.com/product/4116)

### Example Devices

- [Skeleton Device](https://github.com/merliot/skeleton) (template for new devices)
- [Device Hub](https://github.com/merliot/hub) (a hub is a device also)
- [Relay Controller](https://github.com/merliot/relays)
- [Garage Door Opener](https://github.com/merliot/garage)
- [MorningStar Solar Charge Controller](https://github.com/merliot/ps30m) (Modbus)

### Making a New Device

## Environment Variables

These variables configure the hub and devices:

#### `BACKUP`

By default, the each device will dial into the hub that created the device.  To also dial into a backup hub, set `BACKUP` to the backup hub's address.  

For example, a primary hub is at local address `http://192.168.1.10`.  Any device created on the primary hub will dial into the primary hub's address.  A backup hub is at cloud address `https://hub.merliot.net`.  Set `BACKUP=https://hub.merliot.net` on the primary hub.  Now the devices created on the primary hub will dial into both hubs.

> [!TIP]
> You can additionally set `BACKUP=http://192.168.1.10` on the backup hub, so regardless of which hub creates the device, the device will dial into both hubs.

> [!IMPORTANT]
> The backup hub must have the same [USER/PASSWD](#user-passwd) and [WIFI](#wifi_ssids-wifi_passphrases) settings as the primary hub.

#### `DEVICES`

Hub devices.  This is a JSON-formatted list of devices.  The format is:

```
{
	"<id>": {
		"Model": "<model>",
		"Name": "<name>",
		"DeployParams": "<deploy params>"
	},
}
```

Example with two devices:

```
{
	"6bb645c9-db12e9c9": {
		"Model": "skeleton",
		"Name": "example",
		"DeployParams": "target=demo\u0026gpio-default=on"
	},
	"6bccaffd-6d8fab72": {
		"Model": "garage",
		"Name": "garage",
		"DeployParams": "target=demo\u0026door=garage+door\u0026relay=DEMO0"
	},
}
```

#### `PORT`

Port the hub listens on, default is `PORT=8000`.

#### `USER, PASSWD`

Set user and password for HTTP Basic Authentication on the hub.  The user will be prompted for user/password when browsing to the hub.  These values (if set) are automatically passed down to the device when deployed, and the device connects to the hub using these creditials.  For example:

- `USER=foo`
- `PASSWD=bar`

#### `WIFI_SSIDS, WIFI_PASSPHRASES`

Set Wifi SSID(s) and passphrase(s) for Wifi-enabled devices built with TinyGo.  These are matched comma-delimited lists.  For each SSID, there should be a matching passphrase.  For example:

- `WIFI_SSIDS="test,backup"`
- `PASSPHRASES="testtest,backdown"`

So testtest goes with SSID test, and backdown goes with SSID backup.

#### `WS_SCHEME`

Websocket scheme to use for dialing back into the hub.  Default is `WS_SCHEME=ws://`.  If the hub is running under `https://`, then set `WS_SCHEME=wss://`.

## Hub Memory and CPU Requirements

The hub consumes little memory (or CPU) and can run on a Linux machine (or virtual machine) with a minimum of 256M and 2G disk space.
