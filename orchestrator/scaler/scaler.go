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

var renew = make(chan int)
var activeServices = make(map[int]core.CaronteService)
var r1 = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewServiceScale() (ServiceScale, error) {

	swarmEngine, err := engine.NewSwarm()

	return ServiceScale{
		SwarmEngine: swarmEngine,
	}, err
}

func (s ServiceScale) Init(service chan core.CaronteService) {

	go func() {
		for {
			select {
			case threadID := <-renew:
				_, contains := activeServices[threadID]
				if contains {
					go s.worker(activeServices[threadID], renew)
				}

			case service := <-service:

				if (service == core.CaronteService{}) {
					activeServices = make(map[int]core.CaronteService)

				} else {
					service.Thread = r1.Intn(1000)
					activeServices[service.Thread] = service
					go s.worker(service, renew)
				}
			}
		}
	}()

}

func (s ServiceScale) worker(service core.CaronteService, renew chan<- int) {

	swarmService, err := s.SwarmEngine.GetService(service.Id)
	if err == nil {
		if time.Now().After(swarmService.Meta.UpdatedAt.Add(time.Duration(service.CoolDown) * time.Second)) {

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
		}
	} else {
		zap.S().Error(err)
	}

	time.Sleep(time.Duration(service.ServiceScheduler) * time.Second)

	renew <- service.Thread

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
						_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
						if err != nil {
							zap.S().Error(err)
						}
					}
				}

			} else if direction == ScaleDirectionDown {

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
		} else if direction == ScaleDirectionDown {

			if instances*service.MaxReplicasPerNode == targetReplicas {
				_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
				if err != nil {
					zap.S().Error(err)
				}
			}
		}
	} else {

		if (direction == ScaleDirectionUp && targetReplicas <= service.Max) ||
			(direction == ScaleDirectionDown && targetReplicas >= service.Min) {
			_, err := s.SwarmEngine.Scale(service.Name, targetReplicas)
			if err != nil {
				zap.S().Error(err)
			}
		}
	}
}
