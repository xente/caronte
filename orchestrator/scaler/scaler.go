package scaler

import (
	"Caronte/core"
	"Caronte/engine"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

const (
	ScaleDirectionUp   = 1
	ScaleDirectionDown = -1
)

type ServiceScale struct {
	SwarmEngine engine.SwarmEngine
}

var renew = make(chan string)
var activeServices = make(map[string]core.CaronteService)
var r1 = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewServiceScale() (ServiceScale, error) {

	swarmEngine, err := engine.NewSwarm()

	return ServiceScale{
		SwarmEngine: swarmEngine,
	}, err
}

func (s ServiceScale) Init(service chan core.CaronteService, unsuscribe chan core.CaronteService) {

	go func() {
		for {
			select {
			case name := <-renew:
				_, contains := activeServices[name]
				if contains {
					go s.worker(activeServices[name], renew)
				}

			case service := <-service:

				if (service == core.CaronteService{}) {
					activeServices = make(map[string]core.CaronteService)

				} else {
					service.Thread = r1.Intn(1000)
					service.InstanceSpecs.UpdatedAt = activeServices[service.Name].InstanceSpecs.UpdatedAt
					zap.S().Infof("Service %s subscribed", service.Name)
					activeServices[service.Name] = service
					go s.worker(service, renew)
				}

			case unsuscribe := <-unsuscribe:
				zap.S().Infof("Service %s unsuscribed", unsuscribe.Name)
				delete(activeServices, unsuscribe.Name)
			}
		}
	}()

}

func (s ServiceScale) worker(service core.CaronteService, renew chan<- string) {

	result, err := service.MetricProvider.Query(service.MetricSpecs)
	if err != nil {
		zap.S().Error(err)
	} else {

		if result >= service.ScaleUpThreshold {
			s.Scale(service, ScaleDirectionUp)
		} else if result <= service.ScaleDownThreshold {
			s.Scale(service, ScaleDirectionDown)
		}

	}
	time.Sleep(time.Duration(service.ServiceScheduler) * time.Second)

	renew <- service.Name

}

func (s ServiceScale) Scale(service core.CaronteService, direction int) {

	total, err := s.SwarmEngine.TotalActiveTasks(service.Id)
	if err != nil {
		zap.S().Error(err)
	}

	targetReplicas := total + (service.Step * direction)
	if targetReplicas < service.Min {
		targetReplicas = service.Min
	} else if targetReplicas > service.Max {
		targetReplicas = service.Max
	}

	if service.InstanceSpecs.Provider != "" {
		instances := service.InstanceProvider.RunningInstances(service.InstanceSpecs)
		zap.S().Debugf("%d - TargetReplicas %d , Instances %d, active %d ", service.Thread, targetReplicas, instances, total)
		if (targetReplicas / service.MaxReplicasPerNode) != instances {
			if direction == ScaleDirectionUp {
				pending, _ := s.SwarmEngine.PendingTasks(service.Id)
				//Run Infrastructure scale up only when there are not pending tasks
				if pending == 0 {
					ready := service.InstanceProvider.Scale(service.InstanceSpecs, ScaleDirectionUp)

					if ready {
						service.InstanceSpecs.UpdatedAt = time.Now().Add(time.Duration(service.InstanceSpecs.CoolDown) * time.Second)
						//Override service to store the UpadtedAt value
						activeServices[service.Name] = service
						zap.S().Debugf("%d - Instance UpdatedAt ", service.Thread, service.InstanceSpecs.UpdatedAt)
						_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
						if err != nil {
							zap.S().Error(err)
						}
					}
				}

			} else if direction == ScaleDirectionDown {

				if time.Now().After(service.InstanceSpecs.UpdatedAt) {
					pending, _ := s.SwarmEngine.PendingTasks(service.Id)
					//Run Infrastructure scale down only when there are not pending tasks
					if pending == 0 {
						service.InstanceProvider.Scale(service.InstanceSpecs, ScaleDirectionDown)
					}

					if instances*service.MaxReplicasPerNode == targetReplicas {

						_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
						if err != nil {
							zap.S().Error(err)
						}
					}

				}

			}
		} else if direction == ScaleDirectionDown {
			if time.Now().After(service.InstanceSpecs.UpdatedAt) {
				if instances*service.MaxReplicasPerNode == targetReplicas {
					_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
					if err != nil {
						zap.S().Error(err)
					}
				}
			}
		}
	} else {

		zap.S().Debugf("%d - TargetReplicas %d ,  active %d ", service.Thread, targetReplicas, total)
		if direction == ScaleDirectionUp && targetReplicas <= service.Max {
			_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
			if err != nil {
				zap.S().Error(err)
			}
			service.UpdatedAt = time.Now().Add(time.Duration(service.ServiceCoolDownDelay) * time.Second)
			activeServices[service.Name] = service
			zap.S().Debugf("%d - Service UpdatedAt ", service.Thread, service.UpdatedAt)
		}
		if direction == ScaleDirectionDown && targetReplicas >= service.Min {
			if time.Now().After(service.UpdatedAt) {
				_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
				if err != nil {
					zap.S().Error(err)
				}

			}
		}
	}
}
