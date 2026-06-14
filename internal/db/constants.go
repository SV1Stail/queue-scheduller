package db

import "errors"

var (
	errBadRequest = errors.New("bad request")
)

type PostStatus string

const (
	PostStatusDraft      PostStatus = "DRAFT"
	PostStatusScheduled  PostStatus = "SCHEDULED"
	PostStatusPublishing PostStatus = "PUBLISHING"
	PostStatusDone       PostStatus = "DONE"
	PostStatusFailed     PostStatus = "FAILED"
	PostStatusArchived   PostStatus = "ARCHIVED"
)
