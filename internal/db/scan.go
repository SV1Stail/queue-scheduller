package db

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func scanPost(row pgx.Row) (*Post, error) {
	post := &Post{}
	var publishAt time.Time
	err := row.Scan(
		&post.ID,
		&post.PublishChannel,
		&post.Data,
		&post.Status,
		&publishAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Attempts,
	)
	if err != nil {
		return nil, err
	}
	post.PublishAt = timestamppb.New(publishAt)

	return post, nil
}

func scanPosts(rows pgx.Rows) ([]*Post, error) {
	var posts []*Post
	for rows.Next() {
		post := &Post{}
		var publishAt time.Time

		err := rows.Scan(
			&post.ID,
			&post.PublishChannel,
			&post.Data,
			&post.Status,
			&publishAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Attempts,
		)
		if err != nil {
			log.Err(err).Msg("scan failed")
			return nil, err
		}
		post.PublishAt = timestamppb.New(publishAt)

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).Msg("rows error")
		return nil, err
	}

	return posts, nil
}
