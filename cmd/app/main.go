package main

import (
	"cloud-go-testtask/internal/config"
	"cloud-go-testtask/internal/delivery"
	"cloud-go-testtask/internal/repository/cache"
	"cloud-go-testtask/internal/repository/rdbms"
	"cloud-go-testtask/internal/usecase"
	"context"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

/*
In main() function there is:
- init config: cleanenv

- init logger: slog

- init router: chi, render

- run server
*/
func main() {
	cfg := config.MustLoad()
	//config.LoadEnv()

	logger := setupLogger(cfg.Env)

	db, err := sql.Open("postgres", cfg.DBConfig.GetPostgresDSN())
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", "error", err)
		log.Fatalf("Failed to ping database: %v", err)
	}

	rdbmsRepo := rdbms.NewPlaylistRepositoryRDBMS(db)
	cacheRepo := cache.NewPlaylistRepositoryCache()

	defaultPlaylistID := 1
	if repo, ok := rdbmsRepo.(*rdbms.PlaylistRepositoryRDBMS); ok {
		repo.SetDefaultPlaylistID(defaultPlaylistID)
	} else {
		logger.Error("Failed to set default playlist ID: invalid repository type")
		log.Fatalf("Invalid repository type")
	}

	uc := usecase.NewPlaylistUseCase(rdbmsRepo, cacheRepo, logger)

	// Инициализация кеша
	if err := uc.InitCache(); err != nil {
		logger.Error("Failed to initialize cache", "error", err)
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	handler := delivery.NewPlaylistHandler(uc, logger)
	router := delivery.NewRouter(handler)

	// Middleware
	//router.Use(middleware.RequestID)
	//router.Use(middleware.Logger)
	//router.Use(middleware.Recoverer)
	//router.Use(middleware.URLFormat)

	// Init server
	logger.Info("Starting server", slog.String("address", cfg.Address))
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Run server
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server")
		}
	}()

	// Graceful shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shut down the server
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("Server forced to shutdown")
	}
	logger.Info("Server exiting")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
