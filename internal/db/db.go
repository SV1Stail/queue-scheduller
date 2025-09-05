package db

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
	// mu   *sync.Mutex
}

type Post struct {
	ID             string               `json:"id"`
	PublishChannel string               `json:"publish_channel"`
	Data           *PostData            `json:"data,omitempty"`
	Status         PostStatus           `json:"status,omitempty"`
	PublishAt      *timestamp.Timestamp `json:"publish_at"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Attempts       int32                `json:"attempts"`
}

type PostData struct {
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
	PostUrl string `json:"urls,omitempty"`
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

type CreatePostRequest struct {
	ID             string
	PublishChannel string
	Data           *PostData
	PublishAt      *timestamp.Timestamp
}

func (db *DB) CreatePostTx(ctx context.Context, tx pgx.Tx, in *CreatePostRequest) error {
	if in == nil || in.Data == nil || in.PublishAt == nil {
		return errBadRequest
	}

	dataJSONB, err := json.Marshal(in.Data)
	if err != nil {
		return err
	}

	timeNow := time.Now()
	query := `
	INSERT INTO posts (
		"id", "publish_channel", "data",
		"status", "publish_at", "created_at", "updated_at"
	) VALUES ($1, $2 ,$3 ,$4 ,$5 ,$6 ,$7)
	`
	_, err = tx.Exec(ctx, query, in.ID,
		in.PublishChannel, dataJSONB, PostStatusScheduled,
		in.PublishAt.AsTime(), timeNow, timeNow,
	)
	if err != nil {
		return err
	}

	return nil
}

type GetPostRequest struct {
	ID string
}

func (db *DB) GetPostTx(ctx context.Context, tx pgx.Tx, in *GetPostRequest) (*Post, error) {
	if in == nil {
		return nil, errBadRequest
	}

	query := `
	SELECT 
		"id", "publish_channel", "data",
		"status", "publish_at", "created_at", 
		"updated_at", "attempts"
	FROM post
	WHERE "id" = $1
	`
	post, err := scanPost(tx.QueryRow(ctx, query, in.ID))
	if err != nil {
		return nil, err
	}

	return post, nil
}

type UpdatePostRequest struct {
	ID             string
	PublishChannel string
	Data           *PostData
	PublishAt      *timestamp.Timestamp
}

func (db *DB) UpdatePostTx(ctx context.Context, tx pgx.Tx, in *UpdatePostRequest) error {
	if in == nil || in.Data == nil || in.PublishAt == nil {
		return errBadRequest
	}

	dataJSONB, err := json.Marshal(in.Data)
	if err != nil {
		return err
	}

	query := `
	UPDATE posts 
	SET 
		"publish_channel" = $1, "data" = $2,
		"publish_at" = $3 , "updated_at" = NOW()
	WHERE "id" = $4
	`
	_, err = tx.Exec(ctx, query, in.PublishChannel, dataJSONB,
		in.PublishAt.AsTime(), in.ID)

	if err != nil {
		return err
	}

	return nil

}

type DeletePostRequest struct {
	ID string
}

func (db *DB) DeletePostTx(ctx context.Context, tx pgx.Tx, in *DeletePostRequest) error {
	if in == nil {
		return errBadRequest
	}

	query := `
	DELETE FROM post 
	WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, in.ID)
	if err != nil {
		return err
	}

	return nil
}
