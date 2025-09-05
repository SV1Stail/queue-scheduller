package app

import (
	"context"

	"github.com/SV1Stail/queue-scheduller/internal/db"
	publish_post_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/publish_post"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
)

// type App struct {
// 	sendTaskInterval time.Duration // max time for state == PUBLISHING
// 	stopCh           chan struct{}
// }

type QueueScheduler struct {
	queue_scheduler_pb.UnimplementedQueueschedulerServer
	DB *db.DB // создать инициализацию
}

func NewQueueSchedulerService() *QueueScheduler {
	return &QueueScheduler{}
}

func (qs *QueueScheduler) CreatePost(
	ctx context.Context,
	req *publish_post_pb.CreatePostRequest,
) (*publish_post_pb.CreatePostResponse, error) {
	if req == nil ||
		req.GetPublishChannel() == "" ||
		req.GetData() == nil ||
		req.GetPublishAt() == nil {
		return nil, errBadRequest
	}

	// qs.DB.CreatePostTx(ctx, pgx.Tx)

	return &publish_post_pb.CreatePostResponse{}, nil
}

func (qs *QueueScheduler) GetPost(
	ctx context.Context,
	req *publish_post_pb.GetPostRequest,
) (*publish_post_pb.GetPostResponse, error) {
	if req == nil {
		return nil, errBadRequest
	}

	return &publish_post_pb.GetPostResponse{}, nil
}

func (qs *QueueScheduler) UpdatePost(
	ctx context.Context,
	req *publish_post_pb.UpdatePostRequest,
) (*publish_post_pb.UpdatePostResponse, error) {
	if req == nil ||
		req.GetId() == "" ||
		req.GetPublishChannel() == "" ||
		req.GetData() == nil ||
		req.GetPublishAt() == nil {
		return nil, errBadRequest
	}

	return &publish_post_pb.UpdatePostResponse{}, nil
}

func (qs *QueueScheduler) DeletePost(
	ctx context.Context,
	req *publish_post_pb.DeletePostRequest,
) (*publish_post_pb.DeletePostResponse, error) {
	if req == nil {
		return nil, errBadRequest
	}

	return &publish_post_pb.DeletePostResponse{}, nil
}
