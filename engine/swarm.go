package engine

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type SwarmClient struct {
	DockerClient *client.Client
}

type SwarmEngine interface {
	IsLeader(nodeID string) (bool, error)
	ServiceCurrentReplicas(serviceID string) (int, error)
	GetService(serviceID string) (swarm.Service, error)
	GetServices(args filters.Args) ([]swarm.Service, error)
	Scale(serviceID string, target int) (bool, error)
	OnGoingTasks(serviceID string) (int, error)
	PendingTasks(serviceID string) (int, error)
	RunningTasks(serviceID string) (int, error)
	TotalActiveTasks(serviceID string) (int, error)
}

func NewSwarm() (SwarmClient, error) {

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return SwarmClient{}, err
	}

	return SwarmClient{DockerClient: dockerClient}, nil
}

func (p SwarmClient) IsLeader(nodeID string) (bool, error) {

	node, _, err := p.DockerClient.NodeInspectWithRaw(context.Background(), nodeID)
	if err != nil {
		return false, err
	}
	return node.ManagerStatus.Leader, nil
}

func (p SwarmClient) ServiceCurrentReplicas(serviceID string) (int, error) {

	ctx := context.Background()

	service, _, err := p.DockerClient.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return 0, err
	}

	return int(*service.Spec.Mode.Replicated.Replicas), nil

}

func (p SwarmClient) GetService(serviceID string) (swarm.Service, error) {
	ctx := context.Background()

	service, _, err := p.DockerClient.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return swarm.Service{}, err
	}

	return service, nil
}

func (p SwarmClient) GetServices(args filters.Args) ([]swarm.Service, error) {

	services, err := p.DockerClient.ServiceList(context.Background(), types.ServiceListOptions{Filters: args})

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (p SwarmClient) Scale(serviceID string, target int) (bool, error) {

	ctx := context.Background()

	total, err := p.TotalActiveTasks(serviceID)
	if err != nil {
		zap.S().Error(err)
	}

	if total != target {

		service, _, err := p.DockerClient.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{})
		if err != nil {
			return false, err
		}

		zap.S().Infof("Scale service %s from %d to %d replicas", service.Spec.Name, total, target)
		targetScale := uint64(target)
		service.Spec.Mode.Replicated.Replicas = &targetScale

		_, err = p.DockerClient.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, types.ServiceUpdateOptions{})
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil

}
func (p SwarmClient) PendingTasks(serviceID string) (int, error) {
	ctx := context.Background()

	taskFilters := filters.NewArgs()
	taskFilters.Add("service", serviceID)

	tasks, err := p.DockerClient.TaskList(ctx, types.TaskListOptions{
		Filters: taskFilters,
	})
	if err != nil {
		return 0, err
	}
	pending := 0
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStatePending {
			pending++
		}
	}

	return pending, nil
}

func (p SwarmClient) TotalActiveTasks(serviceID string) (int, error) {

	ctx := context.Background()

	taskFilters := filters.NewArgs()
	taskFilters.Add("service", serviceID)

	tasks, err := p.DockerClient.TaskList(ctx, types.TaskListOptions{
		Filters: taskFilters,
	})
	if err != nil {
		return 0, err
	}
	ongoing := 0
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning ||
			task.Status.State == swarm.TaskStatePending ||
			task.Status.State == swarm.TaskStateNew ||
			task.Status.State == swarm.TaskStateAllocated ||
			task.Status.State == swarm.TaskStateAssigned ||
			task.Status.State == swarm.TaskStateAccepted ||
			task.Status.State == swarm.TaskStatePreparing ||
			task.Status.State == swarm.TaskStateReady ||
			task.Status.State == swarm.TaskStateStarting {
			ongoing++
		}
	}

	return ongoing, nil
}
func (p SwarmClient) OnGoingTasks(serviceID string) (int, error) {

	ctx := context.Background()

	taskFilters := filters.NewArgs()
	taskFilters.Add("service", serviceID)

	tasks, err := p.DockerClient.TaskList(ctx, types.TaskListOptions{
		Filters: taskFilters,
	})
	if err != nil {
		return 0, err
	}
	ongoing := 0
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStatePending ||
			task.Status.State == swarm.TaskStateNew ||
			task.Status.State == swarm.TaskStateAllocated ||
			task.Status.State == swarm.TaskStateAssigned ||
			task.Status.State == swarm.TaskStateAccepted ||
			task.Status.State == swarm.TaskStatePreparing ||
			task.Status.State == swarm.TaskStateReady ||
			task.Status.State == swarm.TaskStateStarting {
			ongoing++
		}
	}

	return ongoing, nil
}

func (p SwarmClient) RunningTasks(serviceID string) (int, error) {

	ctx := context.Background()

	taskFilters := filters.NewArgs()
	taskFilters.Add("service", serviceID)

	tasks, err := p.DockerClient.TaskList(ctx, types.TaskListOptions{
		Filters: taskFilters,
	})
	if err != nil {
		return 0, err
	}
	running := 0
	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning {
			running++
		}
	}

	if running == 0 {
		return running, nil
	}

	return running, nil
}
