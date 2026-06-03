package app

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"time"

	"github.com/SV1Stail/queue-scheduller/internal/db"
	publisher_pb "github.com/SV1Stail/tg-project-protos/gen/go/publisher"
	"github.com/mbranch/safe-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	ctxTimeout   = 5 * time.Second
	timerTimeout = 5 * time.Second
)

func (qs *QueueScheduler) Workers(ctx context.Context) {
	// доделать graceful shutdown

	go qs.runJob(ctx, timerTimeout, qs.ClearJob)
	go qs.runJob(ctx, timerTimeout, qs.ReadyPublish)

	<-qs.stopCh
}

func (qs *QueueScheduler) runJob(
	ctx context.Context,
	timerDuration time.Duration,
	job func(context.Context) error,
) {
	timer := time.NewTimer(timerDuration)

	method := runtime.FuncForPC(reflect.ValueOf(job).Pointer())
	log.Info().Ctx(ctx).Str("method_name", method.Name()).Msg("LOL")
	if method != nil {
		ctx = context.WithValue(ctx, "method", method.Name())
	}
mainloop:
	for {
		select {
		case <-qs.stopCh:
			break mainloop
		case <-ctx.Done():
			log.Err(ctx.Err()).Ctx(ctx).Msg("context done")
			break mainloop
		case <-timer.C:
			func() {
				ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
				defer cancel()

				err := safe.Do(func() (err error) {
					return job(ctx)
				})
				if err != nil {
					var panicErr safe.PanicError
					if errors.As(err, &panicErr) {
						log.Err(err).Ctx(ctx).
							Interface("panic_value", panicErr.Panic()).
							Interface("stack", panicErr.StackTrace()).
							Msg("job panicked")
					}
					log.Err(err).Ctx(ctx).Msg("job failed")
				}
			}()

			timer.Reset(timerDuration)
		}
	}
}

func (qs *QueueScheduler) ClearJob(ctx context.Context) error {
	return qs.DB.Clear(ctx)
}

func (qs *QueueScheduler) ReadyPublish(ctx context.Context) error {
	posts, err := qs.DB.Publish(ctx)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for _, post := range posts {
		g.Go(func() error {
			ctx = context.WithValue(ctx, "post_id", post.ID)
			ctx = context.WithValue(ctx, "attempt", post.Attempts)

			log.Debug().Ctx(ctx).Msg("post publishing process")

			ctx, cancel := context.WithTimeout(ctx, contextTimeOut)
			defer cancel()

			var resp *publisher_pb.PublishNowResponse
			resp, err := qs.PublisherClient.PublishNow(ctx, &publisher_pb.PublishNowRequest{
				Id:             post.ID,
				PublishChannel: post.PublishChannel,
				Data:           convertPostData(post.Data),
				PublishAt:      post.PublishAt,
			})
			if err != nil {
				log.Err(err).Msg("PublishNow failed")
				// reschedule if post failed
				_ = qs.DB.Reschedule(ctx, &db.RescheduleRequest{
					ID: post.ID,
				})

				return nil
			}

			// delete if publish ok
			_ = qs.DB.DeletePostByID(ctx, &db.DeletePostRequest{
				ID: resp.GetId(),
			})

			return nil
		})
	}

	return nil
}

func convertPostData(data *db.PostData) *publisher_pb.PublishPostData {
	if data == nil {
		return nil
	}

	return &publisher_pb.PublishPostData{
		Title: data.Title,
		Body:  data.Body,
	}
}
