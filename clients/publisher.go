package clients

import (
	"time"

	publisher_pb "github.com/SV1Stail/tg-project-protos/gen/go/publisher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PublisherClient struct {
	conn *grpc.ClientConn
	publisher_pb.PublisherClient
}

type PublisherConfig struct {
	Address string        // "publisher:9090" или "localhost:9090"
	Timeout time.Duration // например: 30 * time.Second
}

func NewPublisherClient(cfg *PublisherConfig) (*PublisherClient, error) {
	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &PublisherClient{
		conn:            conn,
		PublisherClient: publisher_pb.NewPublisherClient(conn),
	}, nil
}

func (c *PublisherClient) Close() error {
	return c.conn.Close()
}
