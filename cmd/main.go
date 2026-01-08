package main

import (
	"log"
	"net"

	"github.com/SV1Stail/queue-scheduller/clients"
	app "github.com/SV1Stail/queue-scheduller/internal/gRPC"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"google.golang.org/grpc"
)

func main() {
	grpcServer := grpc.NewServer()

	publisherClient, err := clients.NewPublisherClient(&clients.PublisherConfig{})
	if err != nil {
		log.Fatalf("Failed to create publisher client: %v", err)
	}
	defer publisherClient.Close()

	queueScheduler := app.NewQueueSchedulerService(publisherClient)

	queue_scheduler_pb.RegisterQueueschedulerServer(grpcServer, queueScheduler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
