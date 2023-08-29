## GPS Device

Reports a GPS location in lat/long and shows the location on a map.

![](images/view.webp)

Uses [Grove GPS (Air530)](https://www.seeedstudio.com/Grove-GPS-Air530-p-4584.html) from Seeed Studio.  The GPS module periodically sends NMEA messages over serial.

![](images/air530.webp)

### Wiring

#### Linux x86-64

Connect the Air530 GPS module to a USB-to-UART dongle, such as:

![](images/usb-uart.jpg)

Connect as follows:

| Air530  | USB-UART |
| ------- | ---------|
| Vcc | Vcc |
| Gnc | Gnd |
| Tx | Rx |
| Rx | Tx |

Plug the dongle into a USB port on a Linux system.  You can verify the GPS module output using minicom.  The device is /dev/ttyUSB0, 9600 baud:

```
sudo minicom -D /dev/ttyUSB0 -b 9600
```

#### Raspberry Pi




#### Target Install Instructions

- [Linux x86-64](./README-x86_64.md)
- [Raspberry Pi](README-rpi.md)
- [Arduino Nano Connect rp2040](README-nano_rp2040.md)
