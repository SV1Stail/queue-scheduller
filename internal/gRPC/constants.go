package app

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errBadRequest = status.Errorf(codes.InvalidArgument, "bad request")
)
