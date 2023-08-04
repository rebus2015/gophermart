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
	OrdersAllHandler(w http.ResponseWriter, r *http.Request)
	BalanceHandler(w http.ResponseWriter, r *http.Request)
	WithdrawHandler(w http.ResponseWriter, r *http.Request)
	WithdrawalsAllHandler(w http.ResponseWriter, r *http.Request)
}

type apiMiddleware interface {
	AuthMiddleware(next http.Handler) http.Handler
	UserJSONMiddleware(next http.Handler) http.Handler
	OrderTexMiddleware(next http.Handler) http.Handler
	WithdrawJSONMiddleware(next http.Handler) http.Handler
}

func NewRouter(m apiMiddleware, h apiHandlers) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/user/", func(r chi.Router) {
		r.With(m.UserJSONMiddleware).
			Post("/register", h.UserRegisterHandler)
		r.With(m.UserJSONMiddleware).
			Post("/login", h.UserLoginHandler)
		r.Route("/", func(r chi.Router) {
			r.Use(m.AuthMiddleware)
			r.Get("/withdrawals", h.WithdrawalsAllHandler)
			r.With(m.OrderTexMiddleware).
				Post("/orders", h.UserOrderNewHandler)
			r.Get("/orders", h.OrdersAllHandler)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.BalanceHandler)				
				r.With(m.WithdrawJSONMiddleware).
					Post("/withdraw", h.WithdrawHandler)
			})

		})

	})

	return r
}
