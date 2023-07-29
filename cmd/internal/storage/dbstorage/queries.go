package dbstorage

import "database/sql"

const (
	userAddQuery        string = "select user_add(@login,@hash)" // если вернулся uuid - ok, null - такой есть
	userLoginQuery      string = "select * from user_check(@login)"
	orderAddQuery       string = "select * from order_add(@id, @number, @status, 0)"
	ordersAllQuery      string = "select * from orders_all(@id)"
	balanceGetQuery     string = "select * from balance(@id)"
	withdrawQuery       string = "select * from withdraw(@id,@num,@exp)"
	withdrawalsAllQuery string = "select * from withdrawals_all(@id)"
	accUpdate           string = "select order_update(@num,@status,@acc)"
	ordersAcc           string = "select * from orders_acc()"
)

type dbOrder struct {
	Num      sql.NullInt64
	Status   sql.NullString
	Accrural sql.NullInt64
	Ins      sql.NullTime
}

type dbWdr struct {
	Num     sql.NullInt64
	Expence sql.NullInt64
	Ins     sql.NullTime
}
