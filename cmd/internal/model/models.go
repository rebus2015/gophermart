package model

import (
	"encoding/json"
	"time"
)

type User struct {
	Id       string `json:"id,omitempty"`   //uuid пользователя
	Login    string `json:"login"`          //login
	Password string `json:"password"`       //login
	Hash     string `json:"hash,omitempty"` //hash for password
}

type Order struct {
	UserId   string    `json:"userid,omitempty"`      //uuid пользователя
	Num      *int64    `json:"number"`                //номер заказа
	Status   string    `json:"status"`                //статус заказа
	Accrural *int64    `json:"accrual,omitempty"`     //начислено баллов лояльности
	Ins      time.Time `json:"uploaded_at,omitempty"` //дата совершения
}

func (o *Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UserId   string `json:"userid,omitempty"`
		Num      *int64 `json:"number"`
		Status   string `json:"status"`
		Accrural *int64 `json:"accrual,omitempty"`
		Ins      string `json:"uploaded_at,omitempty"`
	}{
		UserId:   o.UserId,
		Num:      o.Num,
		Status:   o.Status,
		Accrural: o.Accrural,
		Ins:      o.Ins.Format(time.RFC3339),
	})
}

type Withdraw struct {
	UserId  string    `json:"userid,omitempty"`       //uuid пользователя
	Num     *int64    `json:"order"`                  //номер заказа
	Expence *int64    `json:"sum"`                    //сумма списания баллов
	Ins     time.Time `json:"processed_at,omitempty"` //дата совершения
}

func (w *Withdraw) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UserId  string `json:"userid,omitempty"`
		Num     *int64 `json:"order"`
		Expence *int64 `json:"sum"`
		Ins     string `json:"processed_at,omitempty"`
	}{
		UserId:  w.UserId,
		Num:     w.Num,
		Expence: w.Expence,
		Ins:     w.Ins.Format(time.RFC3339),
	})
}

type Balance struct {
	Current *int64 `json:"current"`   //текущий баланс
	Expence *int64 `json:"withdrawn"` //использовано баллов за весь период
}
