package app

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/SV1Stail/queue-scheduller/internal/db"
	publisher_pb "github.com/SV1Stail/tg-project-protos/gen/go/publisher"
	"github.com/jackc/pgx/v5"
	"github.com/mbranch/safe-go"
)

func (qs *QueueScheduler) workers() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go qs.worker(ctx, 5*time.Second, qs.ClearJob)
	go qs.worker(ctx, 5*time.Second, qs.ReadyPublish)

}

func (qs *QueueScheduler) worker(ctx context.Context,
	timeDur time.Duration,
	job func(context.Context) error,
) {
	timer := time.NewTimer(timeDur)
	defer timer.Stop()
	for {
		select {
		case <-qs.stopCh:
			return
		case <-ctx.Done():
			return
		case <-timer.C:
			ctx, cancel := context.WithTimeout(ctx, timeDur)

			err := safe.Do(func() error {
				return job(ctx)
			})
			if err != nil {
				log.Default().
					Printf("failed to start job %s", string(debug.Stack()))
			}

			cancel()
			timer.Reset(timeDur)
		}
	}
}

func (qs *QueueScheduler) ClearJob(ctx context.Context) error {
	err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
		return qs.DB.ClearTx(ctx, tx)
	})

	return err
}

func (qs *QueueScheduler) ReadyPublish(ctx context.Context) error {
	var posts []*db.Post
	err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) (err error) {
		posts, err = qs.DB.PublishTx(ctx, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	sem := make(chan struct{}, 10)
	wg := sync.WaitGroup{}
	for _, post := range posts {
		sem <- struct{}{}
		wg.Add(1)

		go func(post *db.Post) {
			defer wg.Done()
			defer func() {
				<-sem
			}()

			var resp *publisher_pb.PublishNowResponse
			resp, err := qs.PublisherClient.PublishNow(ctx, &publisher_pb.PublishNowRequest{
				Id:             post.ID,
				PublishChannel: post.PublishChannel,
				Data:           convertPostData(post.Data),
				PublishAt:      post.PublishAt,
			})
			if err != nil {
				err := qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
					return qs.DB.RescheduleTx(ctx, tx, &db.RescheduleRequest{
						ID: post.ID,
					})
				})
				if err != nil {
					log.Default().Println("reschedule failed")
				}

				log.Default().Println("publish failed")

				return
			}

			err = qs.DB.WrapWithTransAction(ctx, func(tx pgx.Tx) error {
				return qs.DB.DeletePostByIDTx(ctx, tx, &db.DeletePostRequest{
					ID: resp.GetId(),
				})
			})
			if err != nil {
				log.Default().Println("delete failed")

				return
			}

			return
		}(post)
	}

	wg.Wait()

	return nil
}

func convertPostData(data *db.PostData) *publisher_pb.PublishPostData {
	if data == nil {
		return nil
	}

	return &publisher_pb.PublishPostData{
		Title: data.Title,
		Body:  data.Body,
		// PostUrl: data.PostUrl, когда постим мы не знаем ссылку
	}
}
