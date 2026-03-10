// internal/hdhr/hdhr.go
package hdhr

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"sync"
)

// Config holds HDHomeRun device configuration.
type Config struct {
	DeviceID   string
	TunerCount int
	BaseURL    string
}

// LineupChannel is a single entry in the HDHomeRun lineup.
type LineupChannel struct {
	GuideNumber string
	GuideName   string
	URL         string
}

// Handler serves HDHomeRun discovery endpoints.
type Handler struct {
	cfg    Config
	mu     sync.RWMutex
	lineup []LineupChannel
}

// New creates a Handler.
func New(cfg Config) *Handler {
	return &Handler{cfg: cfg}
}

// SetLineup updates the channel lineup.
func (h *Handler) SetLineup(channels []LineupChannel) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lineup = channels
}

// ServeDiscover handles GET /discover.json.
func (h *Handler) ServeDiscover(w http.ResponseWriter, r *http.Request) {
	disc := map[string]any{
		"FriendlyName":    "xTeVe",
		"Manufacturer":    "xTeVe",
		"ModelNumber":     "HDHR3-US",
		"FirmwareName":    "hdhomerun3_atsc",
		"FirmwareVersion": "20200101",
		"DeviceID":        h.cfg.DeviceID,
		"DeviceAuth":      "test1234",
		"BaseURL":         h.cfg.BaseURL,
		"LineupURL":       h.cfg.BaseURL + "/lineup.json",
		"TunerCount":      h.cfg.TunerCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disc)
}

// ServeLineup handles GET /lineup.json.
func (h *Handler) ServeLineup(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	lineup := make([]map[string]any, 0, len(h.lineup))
	for _, ch := range h.lineup {
		lineup = append(lineup, map[string]any{
			"GuideNumber": ch.GuideNumber,
			"GuideName":   ch.GuideName,
			"URL":         ch.URL,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineup)
}

// ServeLineupStatus handles GET /lineup_status.json.
func (h *Handler) ServeLineupStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"ScanInProgress": 0,
		"ScanPossible":   1,
		"Source":         "Cable",
		"SourceList":     []string{"Cable"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

type deviceXML struct {
	XMLName     xml.Name `xml:"root"`
	SpecVersion struct {
		Major int `xml:"specVersion>major"`
		Minor int `xml:"specVersion>minor"`
	}
	URLBase      string `xml:"URLBase"`
	DeviceType   string `xml:"device>deviceType"`
	FriendlyName string `xml:"device>friendlyName"`
	Manufacturer string `xml:"device>manufacturer"`
	ModelName    string `xml:"device>modelName"`
	UDN          string `xml:"device>UDN"`
}

// ServeDeviceXML handles GET /device.xml.
func (h *Handler) ServeDeviceXML(w http.ResponseWriter, r *http.Request) {
	dev := deviceXML{
		URLBase:      h.cfg.BaseURL,
		DeviceType:   "urn:schemas-upnp-org:device:MediaServer:1",
		FriendlyName: "xTeVe",
		Manufacturer: "xTeVe",
		ModelName:    "xTeVe",
		UDN:          "uuid:" + h.cfg.DeviceID,
	}
	dev.SpecVersion.Major = 1
	dev.SpecVersion.Minor = 0

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(dev)
}
