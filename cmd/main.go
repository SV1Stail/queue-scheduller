package main

import (
	"context"
	"log"
	"net"

	"github.com/SV1Stail/queue-scheduller/clients"
	"github.com/SV1Stail/queue-scheduller/internal/db"
	app "github.com/SV1Stail/queue-scheduller/internal/gRPC"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// доделать graceful shutdown

	grpcServer := grpc.NewServer()

	publisherClient, err := clients.NewPublisherClient(&clients.PublisherConfig{})
	if err != nil {
		log.Fatalf("Failed to create publisher client: %v", err)
	}

	db := db.MustNewDB(ctx)
	queueScheduler := app.NewQueueSchedulerService(publisherClient, db)
	defer queueScheduler.Close()

	queue_scheduler_pb.RegisterQueueschedulerServer(grpcServer, queueScheduler)

	lis, err := net.Listen("tcp", ":8091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go queueScheduler.Workers(ctx)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	grpcServer.GracefulStop()
}
