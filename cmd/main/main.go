package main

import (
	"L0/internal/cache"
	"L0/internal/config"
	"L0/internal/http-server/handlers/handler"
	"L0/internal/http-server/middleware/mwlogger"
	"L0/internal/kafka/consumer"
	"L0/internal/lib/logger/handlers/slogpretty"
	"L0/internal/lib/logger/sl"
	"L0/internal/service"
	"L0/internal/storage/postgres"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting handler service", slog.String("env", cfg.Env))
	log.Debug("Debug messages are enabled")

	storage, err := postgres.InitDB(cfg)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	orderCache := cache.New()

	orderService := service.New(storage, orderCache)

	kafkaConsumer := consumer.NewConsumer(cfg.Kafka, orderService)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go kafkaConsumer.Run(ctx, wg)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Handle("/", http.FileServer(http.Dir("./static")))
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API routes
	router.Route("/api", func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			orderHandler := handler.NewOrderHandler(orderService)

			r.Get("/", orderHandler.ListOrders)         // GET /api/orders
			r.Post("/", orderHandler.CreateOrder)       // POST /api/orders
			r.Get("/{orderUID}", orderHandler.GetOrder) // GET /api/orders/123
		})
	})

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	sign := <-stop

	log.Info("application stopping", slog.String("signal", sign.String()))
	cancel()
	wg.Wait()

	log.Info("application stopped")

	if err := storage.Close(); err != nil {
		log.Error("failed to close database", slog.String("error", err.Error()))
	}

	log.Info("postgres connection closed")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	h := opts.NewPrettyHandler(os.Stdout)

	return slog.New(h)
}
