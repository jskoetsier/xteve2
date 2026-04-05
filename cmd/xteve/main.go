package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"xteve/internal/api"
	"xteve/internal/auth"
	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/hdhr"
	"xteve/internal/source"
	"xteve/internal/ssdp"
	"xteve/internal/storage"
	"xteve/internal/ui"
	"xteve/internal/xepg"
)

var version = "2.1.0"

func main() {
	var (
		port      = flag.Int("port", 34400, "HTTP port")
		configDir = flag.String("config", defaultConfigDir(), "Config directory")
		debug     = flag.Int("debug", 0, "Debug level 0-3")
	)
	flag.Parse()

	if *debug > 0 {
		log.Printf("debug level: %d", *debug)
	}

	store := storage.New(*configDir)
	if err := store.EnsureDirs("cache", "backup", "temp", "img-cache", "img-upload"); err != nil {
		log.Fatalf("storage: %v", err)
	}

	cfg, err := config.Load(store)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	cfg = config.ApplyEnvOverrides(cfg)
	if *port != 34400 {
		cfg.Port = *port
	}

	publicBaseURL := envOrDefault("XTEVE_BASE_URL", fmt.Sprintf("http://localhost:%d", cfg.Port))

	authSvc := auth.New(auth.Config{
		Enabled:      cfg.AuthEnabled,
		PasswordHash: cfg.AuthPassword,
	})

	xepgDB := xepg.NewDB()

	buf := buffer.New(buffer.Config{
		TunerCount: cfg.TunerCount,
		Type:       cfg.BufferType,
		FFmpegPath: cfg.FFmpegPath,
		VLCPath:    cfg.VLCPath,
	})

	hdhrHandler := hdhr.New(hdhr.Config{
		DeviceID:   "xteve1234",
		TunerCount: cfg.TunerCount,
		BaseURL:    publicBaseURL,
	})

	sourceManager := source.NewManager(cfg, xepgDB, hdhrHandler, buf, publicBaseURL)

	apiHandler := api.New(api.Config{
		Storage:       store,
		Settings:      cfg,
		XEPG:          xepgDB,
		Buffer:        buf,
		SourceManager: sourceManager,
		OnSettingsChanged: func(updated config.Settings) {
			sourceManager.UpdateSettings(config.ApplyEnvOverrides(updated))
			if err := sourceManager.RefreshPlaylist(context.Background()); err != nil {
				log.Printf("source: refresh after settings update failed: %v", err)
			}
		},
		OnChannelsChanged: func() {
			sourceManager.SyncLineup()
		},
	})

	mux := http.NewServeMux()

	// HDHomeRun discovery (always public)
	mux.HandleFunc("/discover.json", hdhrHandler.ServeDiscover)
	mux.HandleFunc("/lineup.json", hdhrHandler.ServeLineup)
	mux.HandleFunc("/lineup_status.json", hdhrHandler.ServeLineupStatus)
	mux.HandleFunc("/device.xml", hdhrHandler.ServeDeviceXML)
	mux.HandleFunc("GET /m3u/", sourceManager.ServeM3U)
	mux.HandleFunc("GET /xmltv/", sourceManager.ServeXMLTV)
	mux.HandleFunc("GET /stream/{id}", sourceManager.ServeStream)

	// API (auth-protected)
	mux.Handle("/api/", authSvc.Middleware(apiHandler.Router()))
	mux.Handle("/ws", apiHandler.Hub())

	// Web UI (served from go:embed in production, or local files in dev)
	mux.Handle("/", serveUI())

	addr := ":" + strconv.Itoa(cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := ssdp.Advertise(ctx, ssdp.Config{
			DeviceID: "xteve1234",
			Port:     cfg.Port,
		}); err != nil && ctx.Err() == nil {
			log.Printf("ssdp: %v", err)
		}
	}()

	sourceManager.Start(ctx)

	log.Printf("xTeVe %s listening on %s", version, addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	srv.Shutdown(context.Background())
}

func defaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.xteve"
}

func serveUI() http.Handler {
	return ui.Handler()
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
