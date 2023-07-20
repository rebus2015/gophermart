package model

type User struct {
	Id    string `json:"id"`    //uuid пользователя
	Login string `json:"login"` //login
	Hash  []byte `json:"hash"`  //hash for password
}

type Order struct {
	UserId   string `json:"userid"` //uuid пользователя
	Num      *int64 `json:"number"` //номер заказа
	Status   string `json:"status"` //статус заказа
	Accrural *int64 `json:"acc"`    //начислено баллов лояльности
}

type Withdraw struct {
	UserId  string `json:"userid"`   //uuid пользователя
	Num     *int64 `json:"number"`   //номер заказа
	Expence *int64 `json:"excpence"` //сумма списания баллов
}
