package middleware

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rebus2015/gophermart/cmd/internal/api/keys"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	"github.com/rebus2015/gophermart/cmd/internal/model"
	"github.com/rebus2015/gophermart/cmd/internal/utils"
)

type middlewares struct {
	r repository
	l *logger.Logger
}

type repository interface {
	UserLogin(user *model.User) (*model.User, error)
}

const compressed string = `gzip`

func NewMiddlewares(_r repository, _l *logger.Logger) *middlewares {
	return &middlewares{r: _r, l: _l}
}

func (m *middlewares) BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Add("WWW-Authenticate", `Basic realm="Give username and password"`)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "No basic auth present"}`))
			return
		}
		usr := &model.User{
			Login:    username,
			Password: password,
		}
		expectedUser, err := m.r.UserLogin(usr)
		if err != nil {
			m.l.Error().Err(err).Msgf("failed to get auth params for user:%s", username)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "failed to get auth params for user due to database error"}`))
			return
		}

		if utils.CheckPasswordHash(password, string(expectedUser.Hash)) {
			m.l.Info().Msgf("user '%s' is successfully authorized", username)
			usr.Id = expectedUser.Id
			ctx := context.WithValue(r.Context(), keys.UserContextKey{}, usr)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		m.l.Info().Msgf("user '%s' is NOT authorized", username)
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (m *middlewares) UserJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader
		if r.Header.Get(`Content-Encoding`) == compressed {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				m.l.Printf("Failed to create gzip reader: %v", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Failed to create gzip reader: %v", err.Error())))
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		user := &model.User{}
		decoder := json.NewDecoder(reader)

		if err := decoder.Decode(user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("[UserJSONMiddleware] Failed to Decode gzip user: %v", err.Error())))
			return
		}

		if user.Login == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`"user.Login is empty"`))
			return
		}
		if user.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`"user.Password is empty"`))
			return
		}

		hash, err := utils.HashPassword(user.Password)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`"user.Password is not valid, could not create Hash"`))
			return
		}
		user.Hash = hash
		m.l.Printf("Incoming request Method: %v, Body: %v", r.RequestURI, user.Login)
		ctx := context.WithValue(r.Context(), keys.UserContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middlewares) WithdrawJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader

		user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
		if !ok {
			m.l.Error().Msgf(
				"Error: [UserRegisterHandler] User info not found in context status-'500'",
			)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: [UserRegisterHandler] User info not found in context status-'500'")))
			return
		}

		if r.Header.Get(`Content-Encoding`) == compressed {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				m.l.Printf("Failed to create gzip reader: %v", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Failed to create gzip reader: %v", err.Error())))
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		wdr := &model.Withdraw{}
		decoder := json.NewDecoder(reader)

		if err := decoder.Decode(wdr); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`"WithdrawJSONMiddleware failed to Decode Withdraw"`))
			return
		}
		if wdr.Num == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`"Withdraw Number is empty"`))
			return
		}
		if wdr.Expence == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`"Withdraw Expence is empty"`))
			return
		}
		if !utils.Valid(*wdr.Num) {
			m.l.Debug().Msgf("Error withraw order num format mismatch on Luhn check: %v", wdr.Num)
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(fmt.Sprintf("Error withraw order num format mismatch on Luhn check: %v", wdr.Num)))
			return
		}
		wdr.UserId = user.Id
		m.l.Printf("Incoming request Method: %v, Body: %v", r.RequestURI, user.Login)
		ctx := context.WithValue(r.Context(), keys.WithdrwContextKey{}, wdr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middlewares) OrderTexMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader
		user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
		if !ok {
			m.l.Error().Msgf(
				"Error: [UserRegisterHandler] User info not found in context status-'500'",
			)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"User info not found in context"`))
			return
		}

		if r.Header.Get(`Content-Type`) != "text/plain" {
			m.l.Error().Msg("LuhnCheckMiddleware: Error reading request.Body, supposed 'text/plain' content type")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`LuhnCheckMiddleware: Error reading request.Body, supposed 'text/plain' content type"`))
			return
		}
		if r.Header.Get(`Content-Encoding`) == compressed {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				m.l.Printf("Failed to create gzip reader: %v", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Failed to create gzip reader: %v", err.Error())))
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}
		number, err := io.ReadAll(reader)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Failed to retrieve request body from: %v", r.RequestURI)))
			return
		}
		m.l.Debug().Msgf("Retrieved request body: %v", number)
		orderNum, err := strconv.ParseInt(string(number), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Failed to convert request body to Int64: %s", string(number))))
			return
		}
		if !utils.Valid(orderNum) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf("Error order format mismatch on Luhn check: %v", string(number))))
			return
		}
		order := &model.Order{
			UserId: user.Id,
			Num:    &orderNum,
			Status: "NEW",
		}
		m.l.Info().Msgf("Incoming request Method: %v, Order: %v", r.RequestURI, order)
		ctx := context.WithValue(r.Context(), keys.OrderContextKey{}, order)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
