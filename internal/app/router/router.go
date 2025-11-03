package router

import (
	"log"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Te8va/GophKeeper/internal/app/config"
	"github.com/Te8va/GophKeeper/internal/app/handler"
	"github.com/Te8va/GophKeeper/internal/app/middleware"
	"github.com/Te8va/GophKeeper/internal/app/service"
)

func NewRouter(cfg *config.Config, authService service.AuthorizationServ, saver service.DataSaverServ, getter service.DataGetterServ, deleter service.DataDeleteServ) chi.Router {
	r := chi.NewRouter()

	if err := middleware.Initialize("info"); err != nil {
		log.Println("Failed to initialize middleware:", err)
	}

	r.Use(middleware.WithLogging)

	r.Mount("/user", newAuthRouter(cfg, authService))
	r.Mount("/", newAPIRouter(cfg, saver, getter, deleter))
	return r
}

func newAuthRouter(cfg *config.Config, auth service.AuthorizationServ) chi.Router {
	r := chi.NewRouter()

	authServ := service.NewAuthorization(auth, cfg.JWTKey)
	authHandler := handler.NewAuthorizationHandler(authServ)

	r.Post("/register", authHandler.RegisterHandler)
	r.Post("/login", authHandler.LoginHandler)

	return r
}

func newAPIRouter(cfg *config.Config, saver service.DataSaverServ, getter service.DataGetterServ, deleter service.DataDeleteServ) chi.Router {
	r := chi.NewRouter()

	saveHandler := handler.NewSaverHandler(saver)
	getHandler := handler.NewGetterHandler(getter)
	deleteHandler := handler.NewDeleteHandler(deleter)

	r.Use(middleware.Auth(cfg.JWTKey))

	r.Route("/data", func(r chi.Router) {
		r.Post("/", saveHandler.CreateDataHandler)
		r.Get("/", getHandler.GetAllDataHandler)
		r.Get("/{id}", getHandler.GetDataHandler)
		r.Put("/{id}", saveHandler.UpdateDataHandler)
		r.Delete("/{id}", deleteHandler.DeleteDataHandler)
	})

	return r
}
