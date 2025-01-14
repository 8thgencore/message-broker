package broker

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/8thgencore/message-broker/pkg/pb/broker/v1"
)

// PublishMessage publishes a message to the given queue.
func (i *Implementation) PublishMessage(
	ctx context.Context, req *pb.PublishMessageRequest,
) (*pb.PublishMessageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	if req.QueueName == "" {
		return nil, status.Error(codes.InvalidArgument, "queue name is empty")
	}

	if req.Message == nil {
		return nil, status.Error(codes.InvalidArgument, "message is nil")
	}

	msgID, err := i.brokerService.PublishMessage(ctx, req.QueueName, req.Message.Data)
	if err != nil {
		i.logger.Error("failed to publish message",
			"queue", req.QueueName,
			"error", err,
		)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PublishMessageResponse{
		MessageId: msgID,
	}, nil
}

// Subscribe subscribes to the given queue and streams messages to the client.
func (i *Implementation) Subscribe(req *pb.SubscribeRequest, stream pb.BrokerService_SubscribeServer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}

	if req.QueueName == "" {
		return status.Error(codes.InvalidArgument, "queue name is empty")
	}

	ctx := stream.Context()
	result, err := i.brokerService.Subscribe(ctx, req.QueueName)
	if err != nil {
		i.logger.Error("failed to subscribe",
			"queue", req.QueueName,
			"error", err,
		)
		return status.Error(codes.Internal, err.Error())
	}
	defer i.brokerService.Unsubscribe(result.ID)

	for {
		select {
		case msg := <-result.Messages:
			if err := stream.Send(&pb.Message{
				Id:   msg.ID,
				Data: msg.Data,
			}); err != nil {
				i.logger.Error("failed to send message",
					"queue", req.QueueName,
					"subscriber", result.ID,
					"error", err,
				)

				return status.Error(codes.Internal, err.Error())
			}
		case <-result.Done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
