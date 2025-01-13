package app

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/8thgencore/message-broker/internal/broker"
	pb "github.com/8thgencore/message-broker/pkg/broker/v1"
)

type grpcServer struct {
	pb.UnimplementedBrokerServiceServer
	broker *broker.Broker
	logger *slog.Logger
}

func newGRPCServer(broker *broker.Broker, logger *slog.Logger) pb.BrokerServiceServer {
	return &grpcServer{
		broker: broker,
		logger: logger,
	}
}

func (s *grpcServer) PublishMessage(ctx context.Context, req *pb.PublishMessageRequest) (*pb.PublishMessageResponse, error) {
	msgID, err := s.broker.PublishMessage(ctx, req.QueueName, req.Message.Data)
	if err != nil {
		s.logger.Error("failed to publish message",
			slog.String("queue", req.QueueName),
			slog.Any("error", err),
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublishMessageResponse{
		MessageId: msgID,
	}, nil
}

func (s *grpcServer) Subscribe(req *pb.SubscribeRequest, stream pb.BrokerService_SubscribeServer) error {
	ctx := stream.Context()

	subID, msgCh, done, err := s.broker.Subscribe(ctx, req.QueueName)
	if err != nil {
		s.logger.Error("failed to subscribe",
			slog.String("queue", req.QueueName),
			slog.Any("error", err),
		)
		return status.Error(codes.Internal, err.Error())
	}
	defer s.broker.Unsubscribe(subID)

	for {
		select {
		case msg := <-msgCh:
			if err := stream.Send(&pb.Message{
				Id:   msg.ID,
				Data: msg.Data,
			}); err != nil {
				s.logger.Error("failed to send message",
					slog.String("queue", req.QueueName),
					slog.String("subscriber", subID),
					slog.Any("error", err),
				)
				return status.Error(codes.Internal, err.Error())
			}
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
} 