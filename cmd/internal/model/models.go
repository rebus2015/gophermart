package model

import (
	"encoding/json"
	"strconv"
	"time"
)

type User struct {
	ID       string `json:"id,omitempty"`   //uuid пользователя
	Login    string `json:"login"`          //login
	Password string `json:"password"`       //login
	Hash     string `json:"hash,omitempty"` //hash for password
}

type Order struct {
	UserID   string    `json:"userid,omitempty"`      //uuid пользователя
	Num      *int64    `json:"number"`                //номер заказа
	Status   string    `json:"status"`                //статус заказа
	Accrural *float64  `json:"accrual,omitempty"`     //начислено баллов лояльности
	Ins      time.Time `json:"uploaded_at,omitempty"` //дата совершения
}

type Accrual struct {
	Num      string   `json:"order"`             //номер заказа
	Status   string   `json:"status"`            //статус заказа
	Accrural *float64 `json:"accrual,omitempty"` //начислено баллов лояльности
}

func (o *Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UserID   string   `json:"userid,omitempty"`
		Num      string   `json:"number"`
		Status   string   `json:"status"`
		Accrural *float64 `json:"accrual,omitempty"`
		Ins      string   `json:"uploaded_at,omitempty"`
	}{
		UserID:   o.UserID,
		Num:      strconv.FormatInt(*o.Num, 10),
		Status:   o.Status,
		Accrural: o.Accrural,
		Ins:      o.Ins.Format(time.RFC3339),
	})
}

type Withdraw struct {
	UserID  string    `json:"userid,omitempty"`       //uuid пользователя
	Num     *int64    `json:"order"`                  //номер заказа
	Expence *float64  `json:"sum"`                    //сумма списания баллов
	Ins     time.Time `json:"processed_at,omitempty"` //дата совершения
}

func (w *Withdraw) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UserID  string   `json:"userid,omitempty"`
		Num     string   `json:"order"`
		Expence *float64 `json:"sum"`
		Ins     string   `json:"processed_at,omitempty"`
	}{
		UserID:  w.UserID,
		Num:     strconv.FormatInt(*w.Num, 10),
		Expence: w.Expence,
		Ins:     w.Ins.Format(time.RFC3339),
	})
}

func (w *Withdraw) UnmarshalJSON(data []byte) error {
	type Alias Withdraw
	aux := &struct {
		NumStr string `json:"order"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	num, err := strconv.ParseInt(aux.NumStr, 10, 64)
	if err != nil {
		return err
	}
	w.Num = &num
	return nil
}

type Balance struct {
	Current *float64 `json:"current"`   //текущий баланс
	Expence *float64 `json:"withdrawn"` //использовано баллов за весь период
}
