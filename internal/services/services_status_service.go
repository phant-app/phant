package services

import (
	"context"
	"time"

	appservicesstatus "phant/internal/app/servicesstatus"
	domainservicesstatus "phant/internal/domain/servicesstatus"
	servicesinfra "phant/internal/infra/services"
	"phant/internal/infra/system"
)

type ServicesStatusService struct {
	service *appservicesstatus.Service
}

const servicesStatusTimeout = 5 * time.Second

func NewServicesStatusService() *ServicesStatusService {
	provider := servicesinfra.NewProviderForCurrentOS(system.NewExecRunner())

	return &ServicesStatusService{service: appservicesstatus.NewService(appservicesstatus.Dependencies{
		Platform:         provider.Platform,
		DiscoverServices: provider.DiscoverServices,
	})}
}

func (s *ServicesStatusService) GetServicesStatus() domainservicesstatus.Snapshot {
	ctx, cancel := context.WithTimeout(context.Background(), servicesStatusTimeout)
	defer cancel()

	return s.service.GetSnapshot(ctx)
}
