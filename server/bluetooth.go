package server

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type Connection struct {
	macAddr        string
	channel        uint8
	callback       func([]byte)
	fileDescriptor int
	file           *os.File
	active         bool
}

func NewConnection(mac string, channel int, callback func([]byte)) *Connection {
	return &Connection{
		macAddr:  mac,
		channel:  uint8(channel),
		callback: callback,
	}
}

func (c *Connection) Connect() error {
	macAsBytes, err := macToBdaddrLE(c.macAddr)
	if err != nil {
		return err
	}

	c.fileDescriptor, err = unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return err
	}

	sa := &unix.SockaddrRFCOMM{
		Channel: c.channel,
		Addr:    macAsBytes,
	}

	if err = unix.Connect(c.fileDescriptor, sa); err != nil {
		return fmt.Errorf("connect failed (mac=%s ch=%d): %v", c.macAddr, c.channel, err)
	}

	c.file = os.NewFile(uintptr(c.fileDescriptor), "rfcomm-spp")

	go func() {
		for {
			buf := make([]byte, 128)
			n, err := c.file.Read(buf)
			if err != nil {
				c.active = false
				log.Println(err)
				return
			}
			go c.callback(buf[:n])
		}
	}()

	c.active = true
	return nil
}

func (c *Connection) Disconnect() {
	unix.Close(c.fileDescriptor)
	c.active = false
}

func (c *Connection) Send(message []byte) error {
	if !c.active {
		return fmt.Errorf("device not connected")
	}
	if _, err := c.file.Write(message); err != nil {
		c.active = false
		return err
	}
	return nil
}

// parse MAC like "11:75:58:70:53:FA" into 6 bytes (Bluetooth uses little-endian in sockaddr_rc)
func macToBdaddrLE(mac string) ([6]byte, error) {
	var out [6]byte
	mac = strings.ReplaceAll(mac, ":", "")
	b, err := hex.DecodeString(mac)
	if err != nil || len(b) != 6 {
		return out, fmt.Errorf("invalid mac %q", mac)
	}
	// sockaddr_rc expects bdaddr in little-endian (reversed)
	for i := 0; i < 6; i++ {
		out[i] = b[5-i]
	}
	return out, nil
}
