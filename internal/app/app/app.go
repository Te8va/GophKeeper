package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func (a *App) InitPostgresStorage(ctx context.Context) error {
	a.logger.Infoln("Using PostgreSQL as storage")

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		a.logger.Infow("Server started", "addr", a.cfg.ServerAddress)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		} else {
			serverErr <- nil
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Infoln("Received shutdown signal, shutting down gracefully...")
	case err := <-serverErr:
		if err != nil {
			a.logger.Errorw("Server error", "error", err)
			return err
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Errorw("HTTP server shutdown failed", "error", err)
		return err
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
