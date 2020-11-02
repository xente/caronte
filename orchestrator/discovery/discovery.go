package discovery

import (
	"Caronte/core"
	"Caronte/engine"
	"Caronte/orchestrator/scaler"
	"context"

	"github.com/docker/docker/api/types/filters"
	"go.uber.org/zap"
)

const caronteEnable = "caronte.enable"

var activeServices map[string]core.CaronteService

type Discovery struct {
	SwarmEngine engine.SwarmEngine
}

func GetActiveServices() map[string]core.CaronteService {
	return activeServices
}

func SetActiveServices(service core.CaronteService) {
	if activeServices == nil {
		activeServices = make(map[string]core.CaronteService)
	}
	activeServices[service.Name] = service
}

func NewDiscovery() (Discovery, error) {
	swarmEngine, err := engine.NewSwarm()
	return Discovery{SwarmEngine: swarmEngine}, err
}

func (d Discovery) CaronteServiceDiscovery(ctx context.Context) {

	services, err := d.SwarmEngine.GetServices(filters.NewArgs(filters.KeyValuePair{Key: "label", Value: caronteEnable}))
	if err != nil {
		zap.S().Error(err)
	}

	newServices := make(map[string]core.CaronteService)
	serviceChan := make(chan core.CaronteService)

	serviceScale, err := scaler.NewServiceScale()
	if err == nil {
		serviceScale.Init(serviceChan)
	} else {
		zap.S().Error(err)
	}

	serviceChan <- core.CaronteService{}

	for _, service := range services {

		if service.Spec.Mode.Global == nil {
			service := core.NewCaronteService(service.ID, service.Spec.Name, service.Spec.Annotations)
			newServices[service.Name] = service
			serviceChan <- service
		} else {
			zap.S().Warnf("Global service %s can not be managed by Caronte", service.Spec.Name)
		}
	}
	activeServices = newServices

}
