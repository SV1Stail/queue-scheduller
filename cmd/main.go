package main

import (
	"context"
	"net"
	"os"

	"github.com/SV1Stail/queue-scheduller/clients"
	"github.com/SV1Stail/queue-scheduller/internal/db"
	app "github.com/SV1Stail/queue-scheduller/internal/gRPC"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	publisherClient, err := clients.NewPublisherClient(&clients.PublisherConfig{})
	if err != nil {
		log.Err(err).Ctx(ctx).Msg("Failed to create publisher client")
		os.Exit(1)
	}

	db := db.MustNewDB(ctx)
	queueScheduler := app.NewQueueSchedulerService(publisherClient, db)
	defer queueScheduler.Close()

	grpcServer := grpc.NewServer()
	queue_scheduler_pb.RegisterQueueschedulerServer(grpcServer, queueScheduler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+queueScheduler.Port)
	if err != nil {
		log.Err(err).Ctx(ctx).Msg("failed to listen")
		os.Exit(1)
	}

	go queueScheduler.Workers(ctx)

	if err := grpcServer.Serve(lis); err != nil {
		log.Err(err).Ctx(ctx).Msg("failed to serve")
		os.Exit(1)
	}

	grpcServer.GracefulStop()
}
