package storage

import "database/sql"

const (
	userAddQuery   string = "select user_add(@login,@hash)" // если вернулся uuid - ok, null - такой есть
	userLoginQuery string = "select * from user_check(@login)"
	orderAddQuery  string = "select * from order_add(@id, @number, @status, 0)"
	ordersAllQuery string = "select * from orders_all(@id)"
)

type dbOrder struct {
	UserId   sql.NullString
	Num      sql.NullInt64
	Status   sql.NullString
	Accrural sql.NullInt64
	Ins      sql.NullTime
}
