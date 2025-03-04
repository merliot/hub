//go:build tinygo

package ntp

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"time"
)

const NTP_PACKET_SIZE = 48

var ntpHost string = "0.pool.ntp.org:123"

func sendNTPpacket(conn net.Conn) error {
	var request = [48]byte{
		0xe3,
	}

	_, err := conn.Write(request[:])
	return err
}

func parseNTPpacket(r []byte) time.Time {
	// the timestamp starts at byte 40 of the received packet and is four bytes,
	// this is NTP time (seconds since Jan 1 1900):
	t := uint32(r[40])<<24 | uint32(r[41])<<16 | uint32(r[42])<<8 | uint32(r[43])
	const seventyYears = 2208988800
	return time.Unix(int64(t-seventyYears), 0)
}

func getCurrentTime(conn net.Conn) (time.Time, error) {

	var response = make([]byte, NTP_PACKET_SIZE)

	if err := sendNTPpacket(conn); err != nil {
		return time.Time{}, err
	}

	n, err := conn.Read(response)
	if err != nil && err != io.EOF {
		return time.Time{}, err
	}
	if n != NTP_PACKET_SIZE {
		return time.Time{},
			fmt.Errorf("expected NTP packet size of %d: %d",
				NTP_PACKET_SIZE, n)
	}

	return parseNTPpacket(response), nil
}

func SetSystemTime() error {

	println("dialing", ntpHost)
	conn, err := net.Dial("udp", ntpHost)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	println("getting current time")
	now, err := getCurrentTime(conn)
	if err != nil {
		return err
	}

	println("setting system time", now.String())
	runtime.AdjustTimeOffset(-1 * int64(time.Since(now)))

	return nil
}
