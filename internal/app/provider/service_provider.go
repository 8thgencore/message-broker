package provider

import (
	"log/slog"

	"github.com/8thgencore/message-broker/internal/config"
	brokerDelivery "github.com/8thgencore/message-broker/internal/delivery/broker"
	brokerService "github.com/8thgencore/message-broker/internal/service/broker"
)

// ServiceProvider is the main application struct.
type ServiceProvider struct {
	Config *config.Config
	logger *slog.Logger

	brokerService  *brokerService.Service
	brokerDelivery *brokerDelivery.Implementation
}

// NewServiceProvider creates a new ServiceProvider instance.
func NewServiceProvider(cfg *config.Config, logger *slog.Logger) *ServiceProvider {
	return &ServiceProvider{
		Config: cfg,
		logger: logger,
	}
}

// BrokerService returns the broker service.
func (s *ServiceProvider) BrokerService() *brokerService.Service {
	if s.brokerService == nil {
		s.brokerService = brokerService.NewService(s.Config.Queues)
	}
	return s.brokerService
}

// BrokerDelivery returns the broker delivery.
func (s *ServiceProvider) BrokerDelivery() *brokerDelivery.Implementation {
	if s.brokerDelivery == nil {
		s.brokerDelivery = brokerDelivery.NewImplementation(s.logger, s.BrokerService())
	}
	return s.brokerDelivery
}
