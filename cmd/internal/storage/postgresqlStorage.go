package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib" // init db driver for postgeSQl\
)

type PostgreSQLStorage struct {
	connection *sql.DB
	context    context.Context
}

type dbConfig interface {
	GetDbConnection() string
}

func NewStorage(ctx context.Context, connectionString string) (*PostgreSQLStorage, error) {
	db, err := restoreDB(ctx, connectionString)
	if err != nil {
		return nil, err
	}
	return &PostgreSQLStorage{connection: db, context: ctx}, nil
}

func restoreDB(ctx context.Context, connectionString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Printf("Unable to open connection to database connection:'%v'  error %s", connectionString, err)
		return nil, fmt.Errorf("unable to connect to database because %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		log.Printf("Cannot ping database because %s", err)
		return nil, fmt.Errorf("cannot ping database because %w", err)
	}
	var restoreDBscript = ""
	_, err = db.ExecContext(ctx, restoreDBscript)
	if err != nil {
		log.Printf("Fail to invoke %s: %v", restoreDBscript, err)
		return nil, fmt.Errorf("fail to invoke %s: %w", restoreDBscript, err)
	}
	return db, nil
}
