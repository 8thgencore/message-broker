package broker

import (
	"log/slog"

	"github.com/8thgencore/message-broker/internal/service"
	pb "github.com/8thgencore/message-broker/pkg/pb/broker/v1"
)

type Implementation struct {
	pb.UnimplementedBrokerServiceServer
	logger        *slog.Logger
	brokerService service.BrokerService
}

func NewImplementation(logger *slog.Logger, brokerService service.BrokerService) *Implementation {
	return &Implementation{
		logger:        logger,
		brokerService: brokerService,
	}
}
