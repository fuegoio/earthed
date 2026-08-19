// Package main implements the Earthed API server entry point.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/joho/godotenv"

	"github.com/fuegoio/earthed/go/api/internal/api"
	"github.com/fuegoio/earthed/go/api/internal/auth"
	"github.com/fuegoio/earthed/go/api/internal/config"
	"github.com/fuegoio/earthed/go/api/internal/cors"
	"github.com/fuegoio/earthed/go/api/internal/httplog"
	"github.com/fuegoio/earthed/go/api/internal/logging"
	"github.com/fuegoio/earthed/go/api/internal/migrations"
	"github.com/fuegoio/earthed/go/api/internal/reader/fetcher"
	"github.com/fuegoio/earthed/go/api/internal/reader/processor"
	"github.com/fuegoio/earthed/go/api/internal/scheduler"
	"github.com/fuegoio/earthed/go/api/internal/store"
	"github.com/fuegoio/earthed/go/api/internal/worker"

	_ "github.com/lib/pq"
)

func main() {
	code, err := run()
	if err != nil {
		log.Fatal(err)
	}
	if code != 0 {
		os.Exit(code)
	}
}

func run() (int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrateOnly := flag.Bool("migrate", false, "Run migrations and exit")
	dumpOpenAPI := flag.Bool("openapi", false, "Print OpenAPI spec as JSON and exit")
	flag.Parse()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("env: %v", err)
	}

	// The OpenAPI spec is derived from huma operations + struct tags alone.
	// Short-circuit before any DB or auth dependency so --openapi works without
	// a running Postgres or a LIMEN_SECRET.
	if *dumpOpenAPI {
		humaMux := http.NewServeMux()
		humaConfig := huma.DefaultConfig("Earthed API", "1.0.0")
		humaConfig.Servers = []*huma.Server{{URL: ""}}
		humaConfig.Tags = api.OpenAPITags()
		humaRouter := humago.New(humaMux, humaConfig)

		apiHandler := api.New(humaRouter, nil, nil, nil, nil)
		apiHandler.RegisterRoutes()

		b, err := humaRouter.OpenAPI().MarshalJSON()
		if err != nil {
			return 0, fmt.Errorf("marshal openapi: %w", err)
		}
		fmt.Println(string(b))
		return 0, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return 0, fmt.Errorf("config: %w", err)
	}

	if _, err := logging.Init(cfg.LogFormat, os.Stderr); err != nil {
		return 0, fmt.Errorf("logging: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return 0, fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return 0, fmt.Errorf("ping db: %w", err)
	}

	if err := migrations.Run(db); err != nil {
		return 0, fmt.Errorf("migrate: %w", err)
	}

	if *migrateOnly {
		log.Println("migrations complete")
		return 0, nil
	}

	st := store.New(db)
	authInst, err := auth.New(cfg, db, st)
	if err != nil {
		return 0, fmt.Errorf("auth: %w", err)
	}

	humaMux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("Earthed API", "1.0.0")
	humaConfig.Servers = []*huma.Server{{URL: ""}}
	humaConfig.Tags = api.OpenAPITags()
	humaRouter := humago.New(humaMux, humaConfig)

	f := fetcher.New(cfg.HTTPTimeout, cfg.HTTPMaxBody, "Earthed")
	apiHandler := api.New(humaRouter, st, authInst, cfg, f)
	apiHandler.RegisterRoutes()

	mux := http.NewServeMux()

	mux.Handle("/auth/", authInst.Handler())
	// Public device-flow endpoints (issue + poll) must be reachable without
	// a session; the confirm + status endpoints sit behind the middleware
	// because they require an authenticated user to approve the grant.
	for _, p := range api.PublicDevicePaths {
		mux.Handle(p, humaMux)
	}
	mux.Handle("/auth/device/confirm", authInst.Middleware(humaMux))
	mux.Handle("/auth/device/status", authInst.Middleware(humaMux))
	mux.Handle("/v1/health", humaMux)
	mux.Handle("/", authInst.Middleware(humaMux))
	mux.Handle("/docs", humaRouter.Adapter())
	mux.Handle("/openapi.json", humaRouter.Adapter())

	if !cfg.DisableSched {
		proc := processor.New(st, f)
		pool := worker.New(proc, cfg.WorkerPool)
		sched := scheduler.New(st, pool, cfg.PollingFreq, cfg.BatchSize)
		go sched.Start(ctx)
	}

	log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httplog.Middleware(cors.Middleware(cfg.TrustedOrigins)(mux)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return 0, fmt.Errorf("server: %w", err)
	}
	return 0, nil
}
