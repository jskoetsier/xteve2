// internal/ssdp/ssdp.go
package ssdp

import (
	"context"
	"fmt"
	"log"

	gossdp "github.com/koron/go-ssdp"
)

// Config holds SSDP advertisement configuration.
type Config struct {
	DeviceID string
	Port     int
}

// Advertise starts an SSDP advertisement and blocks until ctx is cancelled.
func Advertise(ctx context.Context, cfg Config) error {
	usn := fmt.Sprintf("uuid:%s::upnp:rootdevice", cfg.DeviceID)
	location := fmt.Sprintf("http://localhost:%d/device.xml", cfg.Port)

	ad, err := gossdp.Advertise(
		"upnp:rootdevice",
		usn,
		location,
		"xTeVe",
		1800,
	)
	if err != nil {
		return fmt.Errorf("ssdp: %w", err)
	}

	log.Printf("ssdp: advertising on LAN (device %s)", cfg.DeviceID)

	<-ctx.Done()
	ad.Close()
	return nil
}
