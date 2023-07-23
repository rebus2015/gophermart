package storage

import (
	"context"
	"database/sql"
	"fmt"

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
func (s *PostgreSQLStorage) OrdersNew(order *model.Order) error {
	return nil
}
func (s *PostgreSQLStorage) OrdersAll(user *model.User) (*[]model.Order, error) {
	return new([]model.Order), nil
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
