package router

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	//"github.com/rebus2015/gophermart/cmd/internal/config"
	//"github.com/rebus2015/gophermart/cmd/internal/logger"
	//"github.com/rebus2015/gophermart/cmd/internal/model"
)

type apiHandlers interface {
	UserRegisterHandler(w http.ResponseWriter, r *http.Request)
	UserLoginHandler(w http.ResponseWriter, r *http.Request)
	UserOrderNewHandler(w http.ResponseWriter, r *http.Request)
}

type apiMiddleware interface {
	BasicAuthMiddleware(next http.Handler) http.Handler
	UserJSONMiddleware(next http.Handler) http.Handler
	LuhnCheckMiddleware(next http.Handler) http.Handler
}

func NewRouter(m apiMiddleware, h apiHandlers) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Compress(gzip.BestSpeed, contentTypes...))
	// r.Get("/", GetAllHandler(metricStorage))
	// r.Get("/ping", GetDBConnState(postgreStorage))

	r.Route("/api/user/", func(r chi.Router) {
		r.Use(m.UserJSONMiddleware)
		r.Post("/register", h.UserRegisterHandler)
		r.Post("/login", h.UserLoginHandler)
	})
	r.Route("/api/user/orders", func(r chi.Router) {
		r.With(m.BasicAuthMiddleware).
			With(m.LuhnCheckMiddleware).
			Post("/", h.UserOrderNewHandler)
	})
	return r
}
