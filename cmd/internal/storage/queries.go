package storage

const (
	userAddQuery   string = "select user_add(@login,@hash)" // если вернулся uuid - ok, null - такой есть
	userLoginQuery string = "select * from user_check(@login)"
)
