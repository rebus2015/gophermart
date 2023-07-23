package handlers

import (
	"net/http"
	"time"

	"github.com/rebus2015/gophermart/cmd/internal/api/keys"
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	"github.com/rebus2015/gophermart/cmd/internal/model"
	"github.com/rebus2015/gophermart/cmd/internal/utils"
)

type repository interface {
	UserRegister(user *model.User) (string, error)
	UserLogin(user *model.User) (*model.User, error)
	OrdersAll(user *model.User) (*[]model.Order, error)
	OrdersNew(user *model.User, order *model.Order) error
	Balance(user *model.User, orderNum string) (*model.Balance, error)
	Withdraw(request *model.Withdraw) error
	Withdrawals(user *model.User) (*[]model.Order, error)
}

func NewApi(_repo repository, _log *logger.Logger) *api {
	return &api{repo: _repo, log: _log}
}

type api struct {
	repo repository
	log  *logger.Logger
}

func (a *api) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Printf(
			"Error: [UserRegisterHandler] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}

	id, err := a.repo.UserRegister(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("UserRegisterHandler failed to register, database error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id == "" { //такой уже есть 409
		a.log.Err(err).Msgf("UserRegisterHandler failed, login [%s] is busy", user.Login)
		w.WriteHeader(http.StatusConflict)
		return
	}
	// иначе 200
	w.WriteHeader(http.StatusOK)
	a.log.Info().Msgf("User successfully registered: [%s]", user.Login)
}

func (a *api) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Printf(
			"Error: [UserLoginHandler] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}

	userAcc, err := a.repo.UserLogin(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("UserLoginHandler: failed to log in")
		return
	}
	if userAcc == nil { //такого нет 401
		a.log.Err(err).Msgf("UserLoginHandler: failed, login/pass [%s] failed", user.Login)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pass := utils.CheckPasswordHash(user.Password, string(userAcc.Hash))
	if !pass {
		a.log.Err(err).Msgf("UserLoginHandler: failed, login/pass [%s] hash mismatch", user.Login)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// иначе 200
	w.WriteHeader(http.StatusOK)
	a.log.Info().Msgf("User successfully logged in: [%s]", user.Login)
}

func (a *api) UserOrderNewHandler(w http.ResponseWriter, r *http.Request) {
	orderNum, ok := r.Context().Value(keys.OrderContextKey{}).(*int64)
	if !ok {
		a.log.Printf(
			"Error: [UserOrderNewHandler] Order info not found in context status-'500'",
		)
		http.Error(w, "Order info not found in context", http.StatusInternalServerError)
		return
	}
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Printf(
			"Error: [UserRegisterHandler] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}
	order := model.Order{
		UserId: user.Id,
		Num:    orderNum,
		Status: "NEW",
		Ins:    time.Now(),
	}
	err := a.repo.OrdersNew(user, &order)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("UserRegisterHandler failed to register, database error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	a.log.Info().Msgf("User successfully registered: [%s]", user.Login)
}
