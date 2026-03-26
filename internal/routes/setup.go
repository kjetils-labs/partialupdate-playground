package routes

import (
	"log/slog"
	"net/http"
	"partialupdate/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	Router *chi.Mux
	Store  *database.PersonStore
}

type Routing func(*Router)

func Setup(uri, dbName, collName string, routes ...Routing) (*Router, error) {
	router := &Router{
		Router: chi.NewRouter(),
	}
	router.Router.Use(middleware.Logger)
	router.Router.Use(middleware.Recoverer)

	store, err := database.NewMongoStore(uri, dbName, collName)
	if err != nil {
		return nil, err
	}
	router.Store = store

	for _, route := range routes {
		route(router)
	}
	return router, nil
}

func (r *Router) Start(addr string) {

	chi.Walk(r.Router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Info("route available", "method", method, "route", route, "middlewares", len(middlewares))
		return nil
	})

	slog.Info("listening", "address", addr)
	http.ListenAndServe(addr, r.Router)
}
