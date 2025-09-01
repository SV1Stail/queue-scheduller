package app

import (
	"context"

	publish_post_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/publish_post"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"github.com/google/uuid"
)

// type App struct {
// 	sendTaskInterval time.Duration // max time for state == PUBLISHING
// 	stopCh           chan struct{}
// }

type QueueScheduler struct {
	queue_scheduler_pb.UnimplementedQueueschedulerServer
}

func NewQueueSchedulerService() *QueueScheduler {
	return &QueueScheduler{}
}

func (qs *QueueScheduler) CreatePost(ctx context.Context, req *publish_post_pb.CreatePostRequest) (*publish_post_pb.CreatePostResponse, error) {
	post := &publish_post_pb.PublishPost{
		Id:             uuid.NewString(),
		PublishChannel: req.PublishChannel,
	}
	return nil, nil
}
