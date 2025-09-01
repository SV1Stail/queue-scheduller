package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
	// mu   *sync.Mutex
}

func (db *DB) WrapWithTransAction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		log.Default().Println("tx begin failed")
		return err
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Default().Println("roll back failed")
		}
	}()

	err = fn(tx)
	if err != nil {
		log.Default().Println("fn failed")
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) CreatePostTx(ctx context.Context, tx pgx.Tx) error {
	query := `
	INSERT INTO posts (
		id, publish_channel, data,
		status, publish_at, created_at, updated_at
	) VALUES ($1, $2 ,$3 ,$4 ,$5 ,$6 ,$7)
	`
	tx.QueryRow(ctx, query, id)
	return nil
}
