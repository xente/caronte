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
var serviceChan chan core.CaronteService
var serviceUnsuscribeChan chan core.CaronteService

type Discovery struct {
	SwarmEngine engine.SwarmEngine
}

func GetActiveServices() map[string]core.CaronteService {
	return activeServices
}

func NewDiscovery() (Discovery, error) {
	swarmEngine, err := engine.NewSwarm()
	serviceChan = make(chan core.CaronteService)
	serviceUnsuscribeChan = make(chan core.CaronteService)

	serviceScale, err := scaler.NewServiceScale()
	if err == nil {
		serviceScale.Init(serviceChan, serviceUnsuscribeChan)
	} else {
		zap.S().Error(err)
	}

	serviceChan <- core.CaronteService{}
	return Discovery{SwarmEngine: swarmEngine}, err
}

func (d Discovery) CaronteServiceDiscovery(ctx context.Context) {

	services, err := d.SwarmEngine.GetServices(filters.NewArgs(filters.KeyValuePair{Key: "label", Value: caronteEnable}))
	if err != nil {
		zap.S().Error(err)
	}

	newServices := make(map[string]core.CaronteService)
	for _, dockerService := range services {

		if dockerService.Spec.Mode.Global == nil {
			service := core.NewCaronteService(dockerService.ID, dockerService.Spec.Name, dockerService.Spec.Annotations)

			if !equals(service, activeServices[service.Name]) {
				serviceChan <- service
			}

			newServices[service.Name] = service
		} else {
			zap.S().Warnf("Global service %s can not be managed by Caronte", dockerService.Spec.Name)
		}
	}
	for key := range activeServices {
		_, containes := newServices[key]
		if !containes {
			serviceUnsuscribeChan <- activeServices[key]
		}
	}
	activeServices = newServices

}

func equals(newService core.CaronteService, service core.CaronteService) bool {
	if newService.ServiceScheduler == service.ServiceScheduler &&
		newService.Max == service.Max &&
		newService.Min == service.Min &&
		newService.Step == service.Step &&
		newService.ServiceCoolDownDelay == service.ServiceCoolDownDelay &&
		newService.MaxReplicasPerNode == service.MaxReplicasPerNode &&
		newService.ScaleUpThreshold == service.ScaleUpThreshold &&
		newService.ScaleDownThreshold == service.ScaleDownThreshold &&
		newService.MetricSpecs.Store == service.MetricSpecs.Store &&
		newService.MetricSpecs.Query == service.MetricSpecs.Query &&
		newService.MetricSpecs.PrometheusStore.Address == service.MetricSpecs.PrometheusStore.Address &&
		newService.MetricSpecs.AwsStore.Period == service.MetricSpecs.AwsStore.Period &&
		newService.MetricSpecs.SQSStore.QueueName == service.MetricSpecs.SQSStore.QueueName &&
		newService.InstanceSpecs.Provider == service.InstanceSpecs.Provider &&
		newService.InstanceSpecs.CoolDown == service.InstanceSpecs.CoolDown &&
		newService.InstanceSpecs.Aws.Filters == service.InstanceSpecs.Aws.Filters {
		return true
	}
	return false
}
