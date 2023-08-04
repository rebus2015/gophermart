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
	UserRegister(w http.ResponseWriter, r *http.Request)
	UserLogin(w http.ResponseWriter, r *http.Request)
	UserOrderNew(w http.ResponseWriter, r *http.Request)
	OrdersAll(w http.ResponseWriter, r *http.Request)
	Balance(w http.ResponseWriter, r *http.Request)
	Withdraw(w http.ResponseWriter, r *http.Request)
	WithdrawalsAll(w http.ResponseWriter, r *http.Request)
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
			Post("/register", h.UserRegister)
		r.With(m.UserJSONMiddleware).
			Post("/login", h.UserLogin)
		r.Route("/", func(r chi.Router) {
			r.Use(m.AuthMiddleware)
			r.Get("/withdrawals", h.WithdrawalsAll)
			r.With(m.OrderTexMiddleware).
				Post("/orders", h.UserOrderNew)
			r.Get("/orders", h.OrdersAll)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.Balance)				
				r.With(m.WithdrawJSONMiddleware).
					Post("/withdraw", h.Withdraw)
			})

		})

	})

	return r
}
