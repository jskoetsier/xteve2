// internal/api/ws_test.go
package api_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"xteve/internal/api"
)

func TestWebSocketBroadcast(t *testing.T) {
	hub := api.NewHub()
	go hub.Run()

	server := httptest.NewServer(hub)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	hub.Broadcast([]byte(`{"type":"log","msg":"hello"}`))

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != `{"type":"log","msg":"hello"}` {
		t.Errorf("got %q", msg)
	}
}
