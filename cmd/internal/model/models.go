package model

import "time"

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

type Withdraw struct {
	UserId  string    `json:"userid,omitempty"`      //uuid пользователя
	Num     *int64    `json:"order"`                 //номер заказа
	Expence *int64    `json:"sum"`                   //сумма списания баллов
	Ins     time.Time `json:"uploaded_at,omitempty"` //дата совершения
}

type Balance struct {
	Current *int64 `json:"current"`   //текущий баланс
	Expence *int64 `json:"withdrawn"` //использовано баллов за весь период
}
