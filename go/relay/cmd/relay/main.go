// Package main is the Planetary relay server entrypoint.
//
// The relay:
//   1. Accepts announcements from Planetary instances when users connect AT Proto identities.
//   2. Maintains persistent WebSocket subscriptions to each announced DID's PDS repo stream.
//   3. Aggregates io.planetary.* record events into global counts (followers, shares, feed subs).
//   4. Streams events back to subscribed instances via the subscribeEvents WebSocket endpoint.
//
// Run: relay [--migrate]
//
// Environment variables (see internal/config/config.go):
//   RELAY_HTTP_ADDR        default :9090
//   RELAY_DATABASE_URL     required
//   RELAY_FANOUT_WORKERS   default 50
//   RELAY_RECONNECT_DELAY  default 5s
//   RELAY_EVENT_RETENTION  default 168h (7 days)
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/fuegoio/planetary/go/relay/internal/config"
	"github.com/fuegoio/planetary/go/relay/internal/fanout"
	"github.com/fuegoio/planetary/go/relay/internal/migrations"
	"github.com/fuegoio/planetary/go/relay/internal/server"
	"github.com/fuegoio/planetary/go/relay/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	migrateOnly := flag.Bool("migrate", false, "run migrations and exit")
	flag.Parse()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return err
	}
	if err := migrations.Run(db); err != nil {
		return err
	}
	if *migrateOnly {
		log.Println("migrations complete")
		return nil
	}

	st := store.New(db)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fan := fanout.New(st, cfg.ReconnectDelay)
	go fan.Start(ctx)

	// Background: purge old relay events periodically.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n, err := st.PurgeOldEvents(ctx, cfg.EventRetention)
				if err != nil {
					slog.Warn("relay: purge events", "err", err)
				} else if n > 0 {
					slog.Info("relay: purged old events", "count", n)
				}
			}
		}
	}()

	srv := server.New(st, fan)
	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = httpSrv.Shutdown(shutCtx)
	}()

	slog.Info("relay: listening", "addr", cfg.HTTPAddr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
