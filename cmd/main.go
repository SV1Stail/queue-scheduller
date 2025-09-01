package main

import (
	"log"
	"net"

	app "github.com/SV1Stail/queue-scheduller/internal/gRPC"
	queue_scheduler_pb "github.com/SV1Stail/tg-project-protos/gen/go/queue_scheduler/queue_scheduler"
	"google.golang.org/grpc"
)

func main() {
	queueScheduler := app.NewQueueSchedulerService()

	grpcServer := grpc.NewServer()

	queue_scheduler_pb.RegisterQueueschedulerServer(grpcServer, queueScheduler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
