package db

import (
	"context"
	"encoding/json"
	"log"
	"time"

	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DB struct {
	pool *pgxpool.Pool
}

// TODO: add config
func MustNewDB(ctx context.Context) *DB {
	config, err := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:5432/queue?sslmode=disable")
	if err != nil {
		panic(err)
	}
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(err)
	}

	return &DB{pool: pool}
}

func (db *DB) Close() {
	db.pool.Close()
}

type Post struct {
	ID             string               `json:"id"`
	PublishChannel string               `json:"publish_channel"`
	Data           *PostData            `json:"data,omitempty"`
	Status         PostStatus           `json:"status,omitempty"`
	PublishAt      *timestamp.Timestamp `json:"publish_at"`
	CreatedAt      *timestamp.Timestamp `json:"created_at"`
	UpdatedAt      *timestamp.Timestamp `json:"updated_at"`
	Attempts       int32                `json:"attempts"`
}

type PostData struct {
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	PostsUrls []string `json:"posts_urls,omitempty"`
}

func (db *DB) WrapWithTransAction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
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
	PublishChannel string
	Data           *PostData
	PublishAt      *timestamp.Timestamp
}

func (db *DB) CreatePost(ctx context.Context, in *CreatePostRequest) (*queue_scheduler_pb.Post, error) {
	dataJSONB, err := json.Marshal(in.Data)
	if err != nil {
		return nil, err
	}

	generatedID := uuid.NewString()
	timeNow := time.Now()
	query := `
	INSERT INTO posts (
		"id", "publish_channel", "data",
		"status", "publish_at", "created_at", "updated_at"
	) VALUES ($1, $2 ,$3 ,$4 ,$5, $6, $7)
	`
	_, err = db.pool.Exec(ctx, query, generatedID,
		in.PublishChannel, dataJSONB, PostStatusScheduled,
		in.PublishAt.AsTime(), timeNow, timeNow,
	)
	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.Post{
		Id:             generatedID,
		PublishChannel: in.PublishChannel,
		Data: &queue_scheduler_pb.PublishPostData{
			Title:   in.Data.Title,
			Body:    in.Data.Body,
			PostUrl: in.Data.PostsUrls,
		},
		PublishAt: in.PublishAt,
		CreatedAt: timestamppb.New(timeNow),
		UpdatedAt: timestamppb.New(timeNow),
		// ImagesUrls: []string{},
	}, nil
}

type GetPostRequest struct {
	ID string
}

func (db *DB) GetPost(ctx context.Context, in *GetPostRequest) (*queue_scheduler_pb.Post, error) {
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
	post, err := scanPost(db.pool.QueryRow(ctx, query, in.ID))
	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.Post{
		Id:             post.ID,
		PublishChannel: post.PublishChannel,
		Data: &queue_scheduler_pb.PublishPostData{
			Title:   post.Data.Title,
			Body:    post.Data.Body,
			PostUrl: post.Data.PostsUrls,
		},
		PublishAt: post.PublishAt,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		// ImagesUrls: []string{},
	}, nil
}

type UpdatePostRequest struct {
	ID             string
	PublishChannel string
	Data           *PostData
	PublishAt      *timestamp.Timestamp
}

func (db *DB) UpdatePost(ctx context.Context, in *UpdatePostRequest) (*queue_scheduler_pb.Post, error) {
	if in == nil || in.Data == nil || in.PublishAt == nil {
		return nil, errBadRequest
	}

	dataJSONB, err := json.Marshal(in.Data)
	if err != nil {
		return nil, err
	}

	timeNow := time.Now()
	// TODO: обновлять только не пустые поля
	query := `
	UPDATE posts 
	SET 
		"publish_channel" = $1, 
		"data" = $2,
		"publish_at" = $3 , 
		"updated_at" = $4
	WHERE "id" = $5
	`
	_, err = db.pool.Exec(ctx, query, in.PublishChannel, dataJSONB,
		in.PublishAt.AsTime(), timeNow, in.ID)

	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.Post{
		Id:             in.ID,
		PublishChannel: in.PublishChannel,
		Data: &queue_scheduler_pb.PublishPostData{
			Title:   in.Data.Title,
			Body:    in.Data.Body,
			PostUrl: in.Data.PostsUrls,
		},
		PublishAt: in.PublishAt,
		UpdatedAt: timestamppb.New(timeNow),
		// ImagesUrls: []string{},
	}, nil

}

type DeletePostRequest struct {
	ID string
}

func (db *DB) DeletePostByID(ctx context.Context, in *DeletePostRequest) error {
	if in == nil {
		return errBadRequest
	}

	query := `
	UPDATE post 
	SET 
		status = 'DONE'
		updated_at = Now()
	WHERE "id" = $1
	`
	_, err := db.pool.Exec(ctx, query, in.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) ClearTx(ctx context.Context, tx pgx.Tx) error {
	query := `
	WITH deleted AS (
		DELETE FROM post
		WHERE status = 'DONE'::text::task_state
		LIMIT $1
	)
	SELECT COUNT(*) AS deleted_count
	FROM deleted
	`

	clearLimit := 100
	clearedNumber := -1

	row := tx.QueryRow(ctx, query, clearLimit)
	err := row.Scan(&clearedNumber)
	if err != nil {
		return err
	}

	log.Default().Printf("cleared number %d", clearedNumber)

	return nil
}

func (db *DB) PublishTx(ctx context.Context, tx pgx.Tx) ([]*Post, error) {
	query := `
	UPDATE post
	SET
		state = 'PUBLISHING'
		updated_at = Now()
	WHERE id IN (
		SELECT id 
		FROM post
		WHERE
			state = 'SCHEDULED'  
			AND run_at < NOW()
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	)
	RETURNING 
		id, publish_channel, data,
		status, publish_at, created_at,
		updated_at, attempts
	`
	readyLimit := 100

	rows, err := tx.Query(ctx, query, readyLimit)
	if err != nil {
		return nil, err
	}

	posts, err := scanPosts(rows)
	if err != nil {
		return nil, err
	}

	// log.Default().Printf("affected rows %d", ok.RowsAffected())

	return posts, nil
}

type CreatePostSomeChannelsRequest struct {
	IDs             []string
	PublishChannels []string
	Data            *PostData
	PublishAt       *timestamp.Timestamp
}

func (db *DB) CreatePostSomeChannelsTx(ctx context.Context, tx pgx.Tx, in *CreatePostSomeChannelsRequest) error {
	if in == nil ||
		in.Data == nil ||
		in.PublishAt == nil ||
		len(in.IDs) == 0 ||
		len(in.PublishChannels) == 0 {
		return errBadRequest
	}

	dataJSONB, err := json.Marshal(in.Data)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO posts (
		"id", "publish_channel", "data",
		"status", "publish_at", "created_at", "updated_at"
	) VALUES (unnest($1), unnest($2) ,$3 ,$4 ,$5 ,$6 ,$7)
	`
	timeNow := time.Now()
	_, err = tx.Exec(ctx, query, in.IDs,
		in.PublishChannels, dataJSONB, PostStatusScheduled,
		in.PublishAt.AsTime(), timeNow, timeNow,
	)
	if err != nil {
		return err
	}

	return nil
}

type GetPostsForChannelRequest struct {
	PublishChannel string
}

func (db *DB) GetPostsForChannelTx(ctx context.Context, tx pgx.Tx, in *GetPostsForChannelRequest) ([]*Post, error) {
	if in.PublishChannel == "" {
		return nil, errBadRequest
	}

	query := `
	SELECT
		"id", "publish_channel", "data",
		"status", "publish_at", "created_at",
		"updated_at", "attempts"
	FROM post
	WHERE "publish_channel" = $1
	`
	rows, err := tx.Query(ctx, query, in.PublishChannel)
	if err != nil {
		return nil, err
	}

	posts, err := scanPosts(rows)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

type RescheduleRequest struct {
	ID string
}

func (db *DB) RescheduleTx(ctx context.Context, tx pgx.Tx, in *RescheduleRequest) error {
	query := `
	UPDATE post
	SET 
		status = CASE
			WHEN attempts >= 3 THEN 'FAILED'
			ELSE 'SCHEDULED'
		END,
		attempts = attempts + 1
		updated_at = Now()
	WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, in.ID)
	if err != nil {
		return err
	}

	return nil
}
