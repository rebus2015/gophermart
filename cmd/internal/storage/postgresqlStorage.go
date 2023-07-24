package storage

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
	GetDbConnection() string
}

func NewStorage(ctx context.Context, lg *logger.Logger, conf dbConfig) (*PostgreSQLStorage, error) {
	db, err := restoreDB(ctx, lg, conf.GetDbConnection())
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
		return nil, fmt.Errorf("Error log in user:[%v] query '%s' error: %v", user.Login, userAddQuery, err)
	}
	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil || !id.Valid {

		return nil, fmt.Errorf("failed to execute transaction %w", err)
	}
	userAcc := model.User{
		Id:       id.String,
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
		return "", fmt.Errorf("Error register user:[%v] query '%s' error: %v", user.Login, userAddQuery, err)
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
		"id":     order.UserId,
		"number": order.Num,
		"status": order.Status,
	}
	var id sql.NullString
	errg := tx.QueryRowContext(ctx, orderAddQuery, args).Scan(&id)
	if errg != nil {
		pgs.log.Printf("StorageError: failed to add order [%v] for user id [%v], query '%s' error: %v", order.Num, order.UserId, orderAddQuery, err)
		return "", fmt.Errorf("StorageError: failed to add order [%v] for user id [%v], query '%s' error: %v", order.Num, order.UserId, orderAddQuery, err)
	}

	// шаг 4 — сохраняем изменения
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction %w", err)
	}
	return id.String, nil
}

func (pgs *PostgreSQLStorage) OrdersAll(user *model.User) (*[]model.Order, error) {
	ctx, cancel := context.WithTimeout(pgs.context, time.Second*5)
	defer cancel()
	args := pgx.NamedArgs{
		"id": user.Id,
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
			mo.Accrural = &o.Accrural.Int64
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

func (s *PostgreSQLStorage) Balance(user *model.User, orderNum string) (*model.Balance, error) {
	var b = model.Balance{}
	return &b, nil
}
func (s *PostgreSQLStorage) Withdraw(request *model.Withdraw) error {
	return nil
}
func (s *PostgreSQLStorage) Withdrawals(user *model.User) (*[]model.Order, error) {
	return new([]model.Order), nil
}
