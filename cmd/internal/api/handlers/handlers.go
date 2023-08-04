package handlers

import (
	"encoding/json"
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
	OrdersNew(order *model.Order) (string, error)
	Balance(user *model.User) (*model.Balance, error)
	Withdraw(request *model.Withdraw) (bool, error)
	Withdrawals(user *model.User) (*[]model.Withdraw, error)
}

type auth interface {
	CreateToken(usr *model.User, expirationTime time.Time) (string, error)
}

func NewAPI(r repository, l *logger.Logger, a auth) *api {
	return &api{repo: r, log: l, auth: a}
}

type api struct {
	repo repository
	log  *logger.Logger
	auth auth
}

func (a *api) UserRegister(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Printf(
			"Error: [UserRegister] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}

	id, err := a.repo.UserRegister(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("UserRegister failed to register, database error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if id == "" { //такой уже есть 409
		a.log.Err(err).Msgf("UserRegister failed, login [%s] is busy", user.Login)
		w.WriteHeader(http.StatusConflict)
		return
	}
	user.ID = id

	expirationTime := time.Now().Add(5 * time.Hour)
	tokenString, err := a.auth.CreateToken(user, expirationTime)
	if err != nil {
		a.log.Err(err).Msgf("UserRegister failed to create JWT toker for login [%s]", user.Login)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// иначе 200
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	w.WriteHeader(http.StatusOK)
	a.log.Info().Msgf("User successfully registered: [%s]", user.Login)
}

func (a *api) UserLogin(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Printf(
			"Error: [UserLogin] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}

	userAcc, err := a.repo.UserLogin(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("UserLogin: failed to log in")
		return
	}
	if userAcc == nil { //такого нет 401
		a.log.Err(err).Msgf("UserLogin: failed, login/pass [%s] failed", user.Login)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pass := utils.CheckPasswordHash(user.Password, string(userAcc.Hash))
	if !pass {
		a.log.Err(err).Msgf("UserLogin: failed, login/pass [%s] hash mismatch", user.Login)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(5 * time.Minute)
	tokenString, err := a.auth.CreateToken(userAcc, expirationTime)
	if err != nil {
		a.log.Err(err).Msgf("UserRegister failed to create JWT toker for login [%s]", user.Login)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// иначе 200
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	// иначе 200
	w.WriteHeader(http.StatusOK)
	a.log.Info().Msgf("User successfully logged in: [%s]", user.Login)
}

func (a *api) UserOrderNew(w http.ResponseWriter, r *http.Request) {
	orderNew, ok := r.Context().Value(keys.OrderContextKey{}).(*model.Order)
	if !ok {
		a.log.Printf(
			"Error: [UserOrderNew] Order info not found in context status-'500'",
		)
		http.Error(w, "Order info not found in context", http.StatusInternalServerError)
		return
	}

	order := model.Order{
		UserID: orderNew.UserID,
		Num:    orderNew.Num,
		Status: orderNew.Status,
		Ins:    time.Now(),
	}
	id, err := a.repo.OrdersNew(&order)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("UserRegister failed to register, database error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	switch id {
	case "":
		{
			w.WriteHeader(http.StatusAccepted)
			a.log.Info().Msgf("Order number [%v] successfully added", *order.Num)
			return
		}
	case order.UserID:
		{
			w.WriteHeader(http.StatusOK)
			a.log.Info().Msgf("Order number [%v] already exists for this user", *order.Num)
		}
	default:
		{
			w.WriteHeader(http.StatusConflict)
			a.log.Warn().Msgf("Order number [%v] is already added by another user", *order.Num)
		}
	}
}

func (a *api) OrdersAll(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		a.log.Error().Msgf(
			"Error: [OrdersAll] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}
	a.log.Info().Msgf("[OrdersAll] Запросим из БД заказы для пользователя: %v", user.ID)
	ordersList, err := a.repo.OrdersAll(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msgf("OrdersAll failed to get orders for user [%v], database error", user.Login)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(*ordersList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		a.log.Info().Msgf("No Orders were found for user [%v]", user.Login)
		return
	}
	a.log.Info().Msgf("[OrdersAll] Получено [%v] записей:", len(*ordersList))
	b, err := json.Marshal(ordersList)
	if err != nil {
		a.log.Err(err).Msgf("ERROR Failed to Marshal ORDERS to JSON: [%+v]", ordersList)
		return
	}
	a.log.Info().Msgf("GOT ORDERS FOM DB: %s", string(b))

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(ordersList)
	if err != nil {
		a.log.Err(err).Msgf("Error: [OrdersAll] Result Json encode error :%v", err)
		http.Error(w, "[OrdersAll] Result Json encode error", http.StatusInternalServerError)
	}
}

func (a *api) Balance(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Error().Msgf(
			"Error: [Balance] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}
	balance, err := a.repo.Balance(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msgf("Balance failed to get balance for user [%v], database error", user.Login)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(balance)
	if err != nil {
		a.log.Err(err).Msgf("Error: [Balance] Result Json encode error :%v", err)
		http.Error(w, "[Balance] Result Json encode error", http.StatusInternalServerError)
	}
	a.log.Debug().Msgf("Возвращаем UpdateJSON result :%v", balance)
}

func (a *api) Withdraw(w http.ResponseWriter, r *http.Request) {
	withdrawNew, ok := r.Context().Value(keys.WithdrwContextKey{}).(*model.Withdraw)
	if !ok {
		a.log.Printf(
			"Error: [Withdraw] Withdraw info not found in context status-'500'",
		)
		http.Error(w, "Withdraw info not found in context", http.StatusInternalServerError)
		return
	}

	withdraw := model.Withdraw{
		UserID:  withdrawNew.UserID,
		Num:     withdrawNew.Num,
		Expence: withdrawNew.Expence,
		Ins:     time.Now(),
	}
	success, err := a.repo.Withdraw(&withdraw)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msg("Withdraw failed to register, database error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !success {
		w.WriteHeader(http.StatusPaymentRequired)
		a.log.Error().Msgf("Withdraw FAIL, order number [%v]. Reason: balance is low.", withdraw.Num)
		return
	}
	w.WriteHeader(http.StatusOK)
	a.log.Info().Msgf("Withdraw order number [%v] successfully added", withdraw.Num)
}

func (a *api) WithdrawalsAll(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(keys.UserContextKey{}).(*model.User)
	if !ok {
		a.log.Error().Msgf(
			"Error: [WithdrawalsAll] User info not found in context status-'500'",
		)
		http.Error(w, "User info not found in context", http.StatusInternalServerError)
		return
	}
	wdrls, err := a.repo.Withdrawals(user)
	if err != nil { //ошибка запроса 500
		a.log.Err(err).Msgf("WithdrawalsAll failed to get Withdrawals for user [%v], database error", user.Login)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(*wdrls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		a.log.Info().Msgf("No Withdrawals were found for user [%v]", user.Login)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(wdrls)
	if err != nil {
		a.log.Err(err).Msgf("Error: [Withdrawals] Result Json encode error :%v", err)
		http.Error(w, "[Withdrawals] Result Json encode error", http.StatusInternalServerError)
	}
	a.log.Debug().Msgf("Возвращаем Withdrawals result :%v", wdrls)
}
