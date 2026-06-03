package app

import (
	"context"
	"sync"

	"github.com/SV1Stail/queue-scheduller/clients"
	"github.com/SV1Stail/queue-scheduller/internal/db"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueueScheduler struct {
	queue_scheduler_pb.UnimplementedQueueschedulerServer
	PublisherClient *clients.PublisherClient
	DB              *db.DB
	stopCh          chan struct{}
	mu              *sync.Mutex
}

func NewQueueSchedulerService(publisherClient *clients.PublisherClient, db *db.DB) *QueueScheduler {
	return &QueueScheduler{
		stopCh:          make(chan struct{}),
		mu:              &sync.Mutex{},
		PublisherClient: publisherClient,
		DB:              db,
	}
}

func (qs *QueueScheduler) Close() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.PublisherClient.Close()
	qs.DB.Close()
	close(qs.stopCh)
}

func (qs *QueueScheduler) CreatePost(
	ctx context.Context,
	req *queue_scheduler_pb.CreatePostRequest,
) (*queue_scheduler_pb.CreatePostResponse, error) {
	if req == nil ||
		req.GetPublishChannel() == "" ||
		req.GetData() == nil ||
		req.GetPublishAt() == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	post, err := qs.DB.CreatePost(ctx, &db.CreatePostRequest{
		PublishChannel: req.PublishChannel,
		Data: &db.PostData{
			Title:     req.GetData().GetTitle(),
			Body:      req.GetData().GetBody(),
			PostsUrls: req.GetData().GetPostUrl(),
		},
		PublishAt: req.PublishAt,
	})

	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.CreatePostResponse{
		Post: post,
	}, nil
}

func (qs *QueueScheduler) UpdatePost(
	ctx context.Context,
	req *queue_scheduler_pb.UpdatePostRequest,
) (*queue_scheduler_pb.UpdatePostResponse, error) {
	if req == nil ||
		req.GetId() == "" ||
		req.GetPublishChannel() == "" ||
		req.GetData() == nil ||
		req.GetPublishAt() == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	post, err := qs.DB.UpdatePost(ctx, &db.UpdatePostRequest{
		ID:             req.GetId(),
		PublishChannel: req.GetPublishChannel(),
		Data: &db.PostData{
			Title:     req.GetData().GetTitle(),
			Body:      req.GetData().GetBody(),
			PostsUrls: req.GetData().GetPostUrl(),
		},
		PublishAt: req.GetPublishAt(),
	})
	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.UpdatePostResponse{
		Post: post,
	}, nil
}

func (qs *QueueScheduler) DeletePost(
	ctx context.Context,
	req *queue_scheduler_pb.DeletePostRequest,
) (*queue_scheduler_pb.DeletePostResponse, error) {
	if req == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	err := qs.DB.DeletePostByID(ctx, &db.DeletePostRequest{
		ID: req.GetId(),
	})
	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.DeletePostResponse{
		Id: req.GetId(),
	}, nil
}

func (qs *QueueScheduler) DeletePostsForChannel(ctx context.Context,
	req *queue_scheduler_pb.DeletePostsForChannelRequest,
) (*queue_scheduler_pb.DeletePostsForChannelResponse, error) {
	return &queue_scheduler_pb.DeletePostsForChannelResponse{}, status.Error(codes.Unimplemented, "not implemented)")
}

func (qs *QueueScheduler) GetPost(
	ctx context.Context,
	req *queue_scheduler_pb.GetPostRequest,
) (*queue_scheduler_pb.GetPostResponse, error) {
	if req == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	post, err := qs.DB.GetPost(ctx, &db.GetPostRequest{
		ID: req.GetId(),
	})
	if err != nil {
		return nil, err
	}

	return &queue_scheduler_pb.GetPostResponse{
		Post: post,
	}, nil
}

func (qs *QueueScheduler) GetPostsForChannel(ctx context.Context,
	req *queue_scheduler_pb.GetPostsForChannelRequest,
) (*queue_scheduler_pb.GetPostsForChannelResponse, error) {
	return &queue_scheduler_pb.GetPostsForChannelResponse{}, status.Error(codes.Unimplemented, "not implemented)")
}
