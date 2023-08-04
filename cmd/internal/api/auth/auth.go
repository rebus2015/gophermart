package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/rebus2015/gophermart/cmd/internal/logger"
	"github.com/rebus2015/gophermart/cmd/internal/model"
)

type auth struct {
	log *logger.Logger
	cfg config
}

func NewAuth(logger *logger.Logger, conf config) *auth {
	return &auth{
		log: logger,
		cfg: conf,
	}
}

type config interface {
	GetSecretKey() []byte
}

type claims struct {
	UserID   string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (a *auth) CreateToken(usr *model.User, expirationTime time.Time) (string, error) {

	key := a.cfg.GetSecretKey()
	claims := &claims{
		UserID:   usr.ID,
		Username: usr.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func (a *auth) CheckToken(tokenString string) (*model.User, error) {
	key := a.cfg.GetSecretKey()
	claims := &claims{}
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, http.ErrAbortHandler
		} else {
			return nil, err
		}
	}
	if !tkn.Valid {
		return nil, http.ErrAbortHandler
	}
	return &model.User{
		ID:    claims.UserID,
		Login: claims.Username,
	}, nil
}
