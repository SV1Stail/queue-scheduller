package app

import (
	"context"
	"sync"

	"github.com/SV1Stail/queue-scheduller/clients"
	"github.com/SV1Stail/queue-scheduller/internal/db"
	publish_post_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/publish_post"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QueueScheduler struct {
	queue_scheduler_pb.UnimplementedQueueschedulerServer
	PublisherClient *clients.PublisherClient
	DB              *db.DB // создать инициализацию
	stopCh          chan struct{}
	mu              *sync.Mutex
}

func NewQueueSchedulerService(publisherClient *clients.PublisherClient) *QueueScheduler {
	return &QueueScheduler{
		stopCh:          make(chan struct{}),
		mu:              &sync.Mutex{},
		PublisherClient: publisherClient,
		// DB: ,
	}
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

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	generatedID := uuid.NewString()

	err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		return qs.DB.CreatePostTx(ctx, tx, &db.CreatePostRequest{
			ID:             generatedID,
			PublishChannel: req.PublishChannel,
			Data: &db.PostData{
				Title:   req.Data.Title,
				Body:    req.Data.Body,
				PostUrl: req.Data.PostUrl,
			},
			PublishAt: req.PublishAt,
		})
	})
	if err != nil {
		return nil, err
	}

	return &publish_post_pb.CreatePostResponse{
		Id: generatedID,
	}, nil
}

func (qs *QueueScheduler) GetPost(
	ctx context.Context,
	req *publish_post_pb.GetPostRequest,
) (*publish_post_pb.GetPostResponse, error) {
	if req == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	var post *db.Post
	var err error
	_ = qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		post, err = qs.DB.GetPostTx(ctx, tx, &db.GetPostRequest{
			ID: req.Id,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return &publish_post_pb.GetPostResponse{
		PublishPost: &publish_post_pb.PublishPost{
			Id:             post.ID,
			PublishChannel: post.PublishChannel,
			Data: &publish_post_pb.PublishPostData{
				Title:   post.Data.Title,
				Body:    post.Data.Body,
				PostUrl: post.Data.PostUrl,
			},
			Status:    string(post.Status),
			PublishAt: post.PublishAt,
			CreatedAt: timestamppb.New(post.CreatedAt),
			UpdatedAt: timestamppb.New(post.UpdatedAt),
		},
	}, nil
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

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		return qs.DB.UpdatePostTx(ctx, tx, &db.UpdatePostRequest{
			ID:             req.Id,
			PublishChannel: req.PublishChannel,
			Data: &db.PostData{
				Title:   req.Data.Title,
				Body:    req.Data.Body,
				PostUrl: req.Data.PostUrl,
			},
			PublishAt: req.PublishAt,
		})
	})
	if err != nil {
		return nil, err
	}

	return &publish_post_pb.UpdatePostResponse{
		Id: req.Id,
	}, nil
}

func (qs *QueueScheduler) DeletePost(
	ctx context.Context,
	req *publish_post_pb.DeletePostRequest,
) (*publish_post_pb.DeletePostResponse, error) {
	if req == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		return qs.DB.DeletePostByIDTx(ctx, tx, &db.DeletePostRequest{
			ID: req.Id,
		})
	})
	if err != nil {
		return nil, err
	}

	return &publish_post_pb.DeletePostResponse{
		Id: req.Id,
	}, nil
}

func (qs *QueueScheduler) GetPostsForChannel(ctx context.Context,
	req *publish_post_pb.GetPostsForChannelRequest,
) (*publish_post_pb.GetPostsForChannelResponse, error) {
	if req.GetPublishChannel() == "" {
		return nil, errBadRequest
	}

	var posts []*db.Post
	var err error
	_ = qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		posts, err = qs.DB.GetPostsForChannelTx(ctx, tx, &db.GetPostsForChannelRequest{
			PublishChannel: req.GetPublishChannel(),
		})
		return err
	})

	var protoPosts []*publish_post_pb.PublishPost
	for _, p := range posts {
		protoPosts = append(protoPosts, &publish_post_pb.PublishPost{
			Id:             p.ID,
			PublishChannel: p.PublishChannel,
			Data: &publish_post_pb.PublishPostData{
				Title:   p.Data.Title,
				Body:    p.Data.Body,
				PostUrl: p.Data.PostUrl,
			},
			Status:    string(p.Status),
			PublishAt: p.PublishAt,
			CreatedAt: timestamppb.New(p.CreatedAt),
			UpdatedAt: timestamppb.New(p.UpdatedAt),
		})
	}

	return &publish_post_pb.GetPostsForChannelResponse{
		PublishPosts: protoPosts,
	}, nil
}

func (qs *QueueScheduler) CreatePostSomeChannels(ctx context.Context,
	req *publish_post_pb.CreatePostSomeChannelsRequest,
) (*publish_post_pb.CreatePostSomeChannelsResponse, error) {
	if req == nil ||
		len(req.GetPublishChannels()) == 0 ||
		req.GetData() == nil ||
		req.GetPublishAt() == nil {
		return nil, errBadRequest
	}

	ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
	defer cancel()

	var IDs []string
	for i := 0; i < len(req.GetPublishChannels()); i++ {
		IDs = append(IDs, uuid.NewString())
	}

	err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		return qs.DB.CreatePostSomeChannelsTx(ctx, tx, &db.CreatePostSomeChannelsRequest{
			IDs:             IDs,
			PublishChannels: req.PublishChannels,
			Data: &db.PostData{
				Title:   req.Data.Body,
				Body:    req.Data.Body,
				PostUrl: req.Data.PostUrl,
			},
			PublishAt: req.PublishAt,
		})
	})
	if err != nil {
		return nil, err
	}

	return &publish_post_pb.CreatePostSomeChannelsResponse{
		Ids: IDs,
	}, nil
}
