package middleware

import (
	"compress/gzip"
	"context"
	"encoding/json"
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
		usr := model.User{
			Login:    username,
			Password: password,
		}
		expectedUser, err := m.r.UserLogin(&usr)
		if err != nil {
			m.l.Error().Err(err).Msgf("failed to get auth params for user:%s", username)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "failed to get auth params for user due to database error"}`))
			return
		}

		if utils.CheckPasswordHash(password, string(expectedUser.Hash)) {
			m.l.Info().Msgf("user '%s' is successfully authorized", username)
			next.ServeHTTP(w, r)
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if user.Login == "" {
			http.Error(w, "user.Login is empty", http.StatusBadRequest)
			return
		}
		if user.Password == "" {
			http.Error(w, "user.Password is empty", http.StatusBadRequest)
			return
		}

		hash, err := utils.HashPassword(user.Password)
		if err != nil {
			http.Error(w, "user.Password is not valid, could not create Hash", http.StatusBadRequest)
			return
		}
		user.Hash = hash
		m.l.Printf("Incoming request Method: %v, Body: %v", r.RequestURI, user.Login)
		ctx := context.WithValue(r.Context(), keys.UserContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middlewares) LuhnCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reader io.Reader
		if r.Header.Get(`Content-Type`)!= "text/plain"{
			m.l.Error().Msg("LuhnCheckMiddleware: Error reading request.Body, supposed 'text/plain' content type")
			http.Error(w, "LuhnCheckMiddleware: Error reading request.Body, supposed 'text/plain' content type",
			http.StatusBadRequest)
			return		
		}		
		if r.Header.Get(`Content-Encoding`) == compressed {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				m.l.Printf("Failed to create gzip reader: %v", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}
		number,err:=io.ReadAll(reader)
		if err!=nil {
			m.l.Debug().Msgf("Failed to retrieve request body from: %v",r.RequestURI)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return 
		}
		m.l.Debug().Msgf("Retrieved request body: %v",number) 
		order,err := strconv.ParseInt(string(number),10,64)
		if err!=nil{
			m.l.Debug().Msgf("Failed to convert request body to Int64: %s",string(number))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return 
		}
		if !utils.Valid(order){
			m.l.Debug().Msgf("Error order format mismatch on Luhn check: %v",order)
			http.Error(w, err.Error(), http.StatusConflict)
			return 
		}
		m.l.Info().Msgf("Incoming request Method: %v, Order: %v", r.RequestURI, order)
		ctx := context.WithValue(r.Context(), keys.OrderContextKey{}, order)
		next.ServeHTTP(w, r.WithContext(ctx))		
	})
}