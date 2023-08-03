package dbstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // init db driver for postgeSQl\
	"github.com/rebus2015/gophermart/cmd/internal/logger"
	"github.com/rebus2015/gophermart/cmd/internal/model"
)

type PostgreSQLStorage struct {
	connection *sql.DB
	context    context.Context
	log        *logger.Logger
}

type dbConfig interface {
	GetDBConnection() string
}

func NewStorage(ctx context.Context, lg *logger.Logger, conf dbConfig) (*PostgreSQLStorage, error) {
	db, err := restoreDB(ctx, lg, conf.GetDBConnection())
	if err != nil {
		return nil, err
	}
	return &PostgreSQLStorage{connection: db, log: lg, context: ctx}, nil
}

func restoreDB(ctx context.Context, log *logger.Logger, connectionString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Err(err).Msgf("Unable to open connection to database connection:'%v'", connectionString)
		return nil, fmt.Errorf("unable to connect to database because %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		log.Err(err).Msgf("Cannot ping database due to error")
		return nil, fmt.Errorf("cannot ping database because %w", err)
	}
	return db, nil
}

func (pgs *PostgreSQLStorage) UserLogin(user *model.User) (*model.User, error) {
	ctx, cancel := context.WithCancel(pgs.context)
	defer cancel()

	tx, err := pgs.connection.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return nil, err
	}
	defer func() {
		rberr := tx.Rollback()
		if rberr != nil {
			pgs.log.Printf("failed to rollback transaction err: %v", rberr)
		}
	}()
	args := pgx.NamedArgs{
		"login": user.Login,
	}

	var id sql.NullString
	var hash []byte
	row := tx.QueryRowContext(ctx, userLoginQuery, args)
	errg := row.Scan(&id, &hash)
	if errg != nil {
		pgs.log.Printf("Error log in user:[%v] query '%s' error: %v", user.Login, userAddQuery, err)
		return nil, fmt.Errorf("error log in user [%v] query '%s' error: %v", user.Login, userAddQuery, err)
	}
	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil || !id.Valid {

		return nil, fmt.Errorf("failed to execute transaction %w", err)
	}
	userAcc := model.User{
		ID:       id.String,
		Login:    user.Login,
		Password: user.Password,
		Hash:     string(hash),
	}

	return &userAcc, nil
}

func (pgs *PostgreSQLStorage) UserRegister(user *model.User) (string, error) {
	ctx, cancel := context.WithCancel(pgs.context)
	defer cancel()

	tx, err := pgs.connection.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return "", err
	}
	defer func() {
		rberr := tx.Rollback()
		if rberr != nil {
			pgs.log.Printf("failed to rollback transaction err: %v", rberr)
		}
	}()
	args := pgx.NamedArgs{
		"login": user.Login,
		"hash":  user.Hash,
	}
	var id sql.NullString
	errg := tx.QueryRowContext(ctx, userAddQuery, args).Scan(&id)
	if errg != nil {
		pgs.log.Printf("Error register user:[%v] query '%s' error: %v", user.Login, userAddQuery, err)
		return "", fmt.Errorf("error register user [%v] query '%s' error: %v", user.Login, userAddQuery, err)
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction %w", err)
	}

	return id.String, nil
}

func (pgs *PostgreSQLStorage) OrdersNew(order *model.Order) (string, error) {
	ctx, cancel := context.WithCancel(pgs.context)
	defer cancel()

	tx, err := pgs.connection.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return "", err
	}
	defer func() {
		rberr := tx.Rollback()
		if rberr != nil {
			pgs.log.Printf("failed to rollback transaction err: %v", rberr)
		}
	}()
	args := pgx.NamedArgs{
		"id":     order.UserID,
		"number": order.Num,
		"status": order.Status,
	}
	var id sql.NullString
	errg := tx.QueryRowContext(ctx, orderAddQuery, args).Scan(&id)
	if errg != nil {
		pgs.log.Printf("StorageError: failed to add order [%v] for user id [%v], query '%s' error: %v", order.Num, order.UserID, orderAddQuery, err)
		return "", fmt.Errorf("storageError. failed to add order [%v] for user id [%v], query '%s' error: %v", order.Num, order.UserID, orderAddQuery, err)
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction %w", err)
	}
	return id.String, nil
}

func (pgs *PostgreSQLStorage) OrdersAll(user *model.User) (*[]model.Order, error) {
	ctx, cancel := context.WithTimeout(pgs.context, time.Second*15)
	defer cancel()
	args := pgx.NamedArgs{
		"id": user.ID,
	}
	rows, err := pgs.connection.QueryContext(ctx, ordersAllQuery, args)
	if err != nil {
		pgs.log.Err(err).Msgf("Error trying to get all orders, query: '%s' error: %v", ordersAllQuery, err)
		return nil, fmt.Errorf("error trying to get all orders, query: '%s' error: %w", ordersAllQuery, err)
	}
	ordersList := new([]model.Order)
	for rows.Next() {
		var o dbOrder
		err = rows.Scan(&o.Num, &o.Status, &o.Accrural, &o.Ins)
		if err != nil {
			pgs.log.Err(err).Msgf("Error trying to Scan Rows error: %v", err)
			return nil, fmt.Errorf("error trying to Scan Rows error: %w", err)
		}
		mo := model.Order{}
		mo.Num = &o.Num.Int64
		mo.Status = o.Status.String
		if o.Accrural.Valid {
			mo.Accrural = &o.Accrural.Float64
		}
		mo.Ins = o.Ins.Time
		*ordersList = append(
			*ordersList, mo)
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}	
	return ordersList, nil
}

func (pgs *PostgreSQLStorage) Balance(user *model.User) (*model.Balance, error) {
	ctx, cancel := context.WithCancel(pgs.context)
	defer cancel()
	tx, err := pgs.connection.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return nil, err
	}
	defer func() {
		rberr := tx.Rollback()
		if rberr != nil {
			pgs.log.Printf("failed to rollback transaction err: %v", rberr)
		}
	}()
	args := pgx.NamedArgs{
		"id": user.ID,
	}

	var balance sql.NullFloat64
	var expence sql.NullFloat64
	row := tx.QueryRowContext(ctx, balanceGetQuery, args)
	errg := row.Scan(&balance, &expence)
	if errg != nil {
		pgs.log.Err(errg).Msgf("[Balance] failed to get balance for user:[%v] error: %v", user.Login, err)
		return nil, fmt.Errorf("[Balance] failed to get balance for user:[%v] error: %v", user.Login, err)
	}
	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil || !expence.Valid {
		return nil, fmt.Errorf("failed to execute transaction %w", err)
	}

	b := model.Balance{
		Current: &balance.Float64,
		Expence: &expence.Float64,
	}
	return &b, nil
}

func (pgs *PostgreSQLStorage) Withdraw(request *model.Withdraw) (bool, error) {
	ctx, cancel := context.WithCancel(pgs.context)
	defer cancel()

	tx, err := pgs.connection.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return false, err
	}
	defer func() {
		rberr := tx.Rollback()
		if rberr != nil {
			pgs.log.Printf("failed to rollback transaction err: %v", rberr)
		}
	}()
	args := pgx.NamedArgs{
		"id":  request.UserID,
		"num": request.Num,
		"exp": request.Expence,
	}
	var result sql.NullBool
	errg := tx.QueryRowContext(ctx, withdrawQuery, args).Scan(&result)
	if errg != nil {
		pgs.log.Printf("StorageError: failed to withdraw [%v] points for user id [%v], query '%s' error: %v", request.Expence, request.UserID, withdrawQuery, err)
		return false, fmt.Errorf("StorageError: failed to withdraw [%v] points for user id [%v], query '%s' error: %v", request.Expence, request.UserID, withdrawQuery, err)
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return false, fmt.Errorf("failed to execute transaction %w", err)
	}
	return result.Bool, nil
}

func (pgs *PostgreSQLStorage) Withdrawals(user *model.User) (*[]model.Withdraw, error) {
	ctx, cancel := context.WithTimeout(pgs.context, time.Second*5)
	defer cancel()
	args := pgx.NamedArgs{
		"id": user.ID,
	}
	rows, err := pgs.connection.QueryContext(ctx, withdrawalsAllQuery, args)
	if err != nil {
		pgs.log.Err(err).Msgf("Error trying to get all orders, query: '%s' error: %v", ordersAllQuery, err)
		return nil, fmt.Errorf("error trying to get all orders, query: '%s' error: %w", ordersAllQuery, err)
	}
	wdrsList := new([]model.Withdraw)
	for rows.Next() {
		var o dbWdr
		err = rows.Scan(&o.Num, &o.Expence, &o.Ins)
		if err != nil {
			pgs.log.Err(err).Msgf("Error trying to Scan Rows error: %v", err)
			return nil, fmt.Errorf("error trying to Scan Rows error: %w", err)
		}
		mo := model.Withdraw{}
		mo.Num = &o.Num.Int64
		mo.Expence = &o.Expence.Float64
		mo.Ins = o.Ins.Time
		*wdrsList = append(
			*wdrsList, mo)
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return wdrsList, nil
}

func (pgs *PostgreSQLStorage) AccruralUpdate(order *model.Order) error {
	ctx, cancel := context.WithCancel(pgs.context)
	defer cancel()

	tx, err := pgs.connection.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return err
	}
	defer func() {
		rberr := tx.Rollback()
		if rberr != nil {
			pgs.log.Printf("failed to rollback transaction err: %v", rberr)
		}
	}()
	args := pgx.NamedArgs{
		"num":    order.Num,
		"status": order.Status,
		"acc":    order.Accrural,
	}

	_, errg := tx.ExecContext(ctx, accUpdate, args)
	if errg != nil {
		pgs.log.Printf("Error AccruralUpdate order num:[%v] query '%s' error: %v", order.Num, accUpdate, err)
		return fmt.Errorf("error AccruralUpdate order num:[%v] query '%s' error: %v", order.Num, accUpdate, err)
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("AccruralUpdate: failed to execute transaction %w", err)
	}
	return nil
}

func (pgs *PostgreSQLStorage) OrdersAcc() ([]model.Order, error) {
	ctx, cancel := context.WithTimeout(pgs.context, time.Second*5)
	defer cancel()

	rows, err := pgs.connection.QueryContext(ctx, ordersAcc)
	if err != nil {
		pgs.log.Err(err).Msgf("Error trying to get all orders, query: '%s' error: %v", ordersAcc, err)
		return nil, fmt.Errorf("error trying to get all orders, query: '%s' error: %w", ordersAcc, err)
	}
	ordersList := new([]model.Order)
	for rows.Next() {
		var o dbOrder
		err = rows.Scan(&o.Num, &o.Status)
		if err != nil {
			pgs.log.Err(err).Msgf("Error trying to Scan Rows error: %v", err)
			return nil, fmt.Errorf("error trying to Scan Rows error: %w", err)
		}
		mo := model.Order{}
		mo.Num = &o.Num.Int64
		mo.Status = o.Status.String
		*ordersList = append(
			*ordersList, mo)
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return *ordersList, nil
}
