package provider

import (
	"log/slog"

	"github.com/8thgencore/message-broker/internal/config"
	brokerDelivery "github.com/8thgencore/message-broker/internal/delivery/broker"
	brokerService "github.com/8thgencore/message-broker/internal/service/broker"
)

type ServiceProvider struct {
	Config *config.Config
	logger *slog.Logger

	brokerService  *brokerService.Service
	brokerDelivery *brokerDelivery.Implementation
}

func NewServiceProvider(cfg *config.Config, logger *slog.Logger) *ServiceProvider {
	return &ServiceProvider{
		Config: cfg,
		logger: logger,
	}
}

func (s *ServiceProvider) BrokerService() *brokerService.Service {
	if s.brokerService == nil {
		s.brokerService = brokerService.NewService(s.Config.Queues)
	}
	return s.brokerService
}

func (s *ServiceProvider) BrokerDelivery() *brokerDelivery.Implementation {
	if s.brokerDelivery == nil {
		s.brokerDelivery = brokerDelivery.NewImplementation(s.logger, s.BrokerService())
	}
	return s.brokerDelivery
}
