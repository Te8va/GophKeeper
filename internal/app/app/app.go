package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/Te8va/GophKeeper/internal/app/config"
	"github.com/Te8va/GophKeeper/internal/app/repository"
	"github.com/Te8va/GophKeeper/internal/app/router"
	"github.com/Te8va/GophKeeper/internal/app/service"
)

type App struct {
	cfg        *config.Config
	logger     *zap.SugaredLogger
	saver      service.DataSaverServ
	getter     service.DataGetterServ
	deleter    service.DataDeleteServ
	auth       service.AuthorizationServ
	httpServer *http.Server
}

func NewApp() (*App, error) {
	cfg := config.NewConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar()
	defer func() {
		if err := logger.Sync(); err != nil {
			sugar.Errorw("Failed to sync logger", "error", err)
		}
	}()

	app := &App{
		cfg:    cfg,
		logger: sugar,
	}

	if err := app.initStorage(); err != nil {
		return nil, err
	}

	app.initServer()
	return app, nil
}

func (a *App) initStorage() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch {
	case a.cfg.DatabaseDSN != "":
		return a.InitPostgresStorage(ctx)
	case a.cfg.FileStoragePath != "":
		return a.initFileStorage()
	default:
		return a.initMemoryStorage()
	}
}

// func (a *App) InitPostgresStorage(ctx context.Context) error {
// 	a.logger.Infoln("Using PostgreSQL as storage")

// 	m, err := migrate.New("file://migrations", a.cfg.DatabaseDSN)
// 	if err != nil {
// 		a.logger.Fatalw("Failed to initialize migrations", "error", err)
// 	}

// 	if err := repository.ApplyMigrations(m); err != nil {
// 		a.logger.Fatalw("Failed to apply migrations", "error", err)
// 	}

// 	pool, err := repository.GetPgxPool(ctx, a.cfg.DatabaseDSN)
// 	if err != nil {
// 		a.logger.Fatalw("Failed to create Postgres connection pool", "error", err)
// 	}

// 	repo := repository.NewDataRepository(pool)

// 	repoAuth := repository.NewAuthorizationRepository(pool)

// 	a.saver = repo
// 	a.getter = repo
// 	a.deleter = repo
// 	a.auth = repoAuth

// 	return nil
// }

func (a *App) InitPostgresStorage(ctx context.Context) error {
	a.logger.Infoln("Using PostgreSQL as storage")

	m, err := migrate.New("file://migrations", a.cfg.DatabaseDSN)
	if err != nil {
		a.logger.Errorw("Failed to initialize migrations", "error", err)
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := repository.ApplyMigrations(m); err != nil {
		a.logger.Errorw("Failed to apply migrations", "error", err)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	pool, err := repository.GetPgxPool(ctx, a.cfg.DatabaseDSN)
	if err != nil {
		a.logger.Errorw("Failed to create Postgres connection pool", "error", err)
		return fmt.Errorf("failed to create Postgres connection pool: %w", err)
	}

	repo := repository.NewDataRepository(pool)
	repoAuth := repository.NewAuthorizationRepository(pool)

	a.saver = repo
	a.getter = repo
	a.deleter = repo
	a.auth = repoAuth

	return nil
}

func (a *App) initFileStorage() error {
	a.logger.Infoln("Using JSON file as storage:", a.cfg.FileStoragePath)

	storage, err := repository.NewFileRepository(a.cfg.FileStoragePath)
	if err != nil {
		a.logger.Fatalw("Failed to initialize JSON repository", "error", err)
	}

	authRepo, err := repository.NewFileAuthRepository(a.cfg.FileAuthStoragePath)
	if err != nil {
		a.logger.Fatalw("Failed to initialize auth repository", "error", err)
	}

	a.saver = storage
	a.getter = storage
	a.deleter = storage
	a.auth = authRepo
	return nil
}

func (a *App) initMemoryStorage() error {
	a.logger.Infoln("Using in-memory storage")
	storage := repository.NewMemoryRepository()
	authRepo := repository.NewMemoryAuthRepository()

	a.saver = storage
	a.getter = storage
	a.deleter = storage
	a.auth = authRepo
	return nil
}

func (a *App) initServer() {
	handler := router.NewRouter(a.cfg, a.auth, a.saver, a.getter, a.deleter)

	a.httpServer = &http.Server{
		Addr:    a.cfg.ServerAddress,
		Handler: handler,
	}
}

func (a *App) Run() error {
	defer func() {
		if err := a.logger.Sync(); err != nil {
			a.logger.Errorw("Failed to sync logger", "error", err)
		}
	}()

	go func() {
		a.logger.Infow("Server started", "addr", a.cfg.ServerAddress)

		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalw("ListenAndServe failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	<-quit

	a.logger.Infoln("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			a.logger.Errorw("HTTP server shutdown failed", "error", err)
		}
	}()

	waitGroupChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitGroupChan)
	}()

	select {
	case <-waitGroupChan:
		a.logger.Infoln("All servers finished cleanly")
	case <-time.After(3 * time.Second):
		a.logger.Warn("Some servers did not finish in time")
	}

	a.logger.Infoln("Server shut down successfully")
	return nil
}

func (a *App) SetConfig(cfg *config.Config) {
	a.cfg = cfg
}

func (a *App) SetLogger(logger *zap.SugaredLogger) {
	a.logger = logger
}

func (a *App) SetSaver(saver service.DataSaverServ) {
	a.saver = saver
}

func (a *App) SetGetter(getter service.DataGetterServ) {
	a.getter = getter
}

func (a *App) SetDeleter(deleter service.DataDeleteServ) {
	a.deleter = deleter
}

func (a *App) SetAuth(auth service.AuthorizationServ) {
	a.auth = auth
}

func (a *App) SetHTTPServer(server *http.Server) {
	a.httpServer = server
}

func (a *App) Config() *config.Config {
	return a.cfg
}

func (a *App) Logger() *zap.SugaredLogger {
	return a.logger
}

func (a *App) Saver() service.DataSaverServ {
	return a.saver
}

func (a *App) Getter() service.DataGetterServ {
	return a.getter
}

func (a *App) Deleter() service.DataDeleteServ {
	return a.deleter
}

func (a *App) Auth() service.AuthorizationServ {
	return a.auth
}

func (a *App) HTTPServer() *http.Server {
	return a.httpServer
}

func (a *App) InitStorage() error {
	return a.initStorage()
}

func (a *App) InitServer() {
	a.initServer()
}
