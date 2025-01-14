package broker

import (
	"log/slog"

	"github.com/8thgencore/message-broker/internal/service"
	pb "github.com/8thgencore/message-broker/pkg/pb/broker/v1"
)

// Implementation is the implementation of the BrokerServiceServer interface.
type Implementation struct {
	pb.UnimplementedBrokerServiceServer
	logger        *slog.Logger
	brokerService service.BrokerService
}

// NewImplementation creates a new Implementation instance.
func NewImplementation(logger *slog.Logger, brokerService service.BrokerService) *Implementation {
	return &Implementation{
		logger:        logger,
		brokerService: brokerService,
	}
}
