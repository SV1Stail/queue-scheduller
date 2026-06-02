package app

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errBadRequest = status.Errorf(codes.InvalidArgument, "bad request")
)

var (
	contextTimeOut time.Duration = 5 * time.Second
)
