package core

import (
	"Caronte/instances"
	"Caronte/metricstores"
	"strconv"

	"github.com/docker/docker/api/types/swarm"
	"go.uber.org/zap"
)

type CaronteService struct {
	Id                 string
	Name               string
	ServiceScheduler   int
	Max                int
	Min                int
	MaxReplicasPerNode int
	Step               int
	ScaleUpThreshold   float64
	ScaleDownThreshold float64
	CoolDown           int
	Thread             int
	MetricSpecs        metricstores.MetricSpecs
	MetricProvider     metricstores.MetricProvider
	InstanceSpecs      instances.ScaleSpecs
	InstanceProvider   instances.InstanceManagerProvider
}

func NewCaronteService(id string, name string, annotations swarm.Annotations) CaronteService {

	serviceScheduler := labelStringToInt(annotations.Labels["caronte.ervice.scheduler.scale.time"])
	max := labelStringToInt(annotations.Labels["caronte.scale.max"])
	min := labelStringToInt(annotations.Labels["caronte.scale.min"])
	step := labelStringToInt(annotations.Labels["caronte.scale.step"])
	coolDown := labelStringToInt(annotations.Labels["caronte.scale.coolDown"])
	maxReplicasPerNode := labelStringToInt(annotations.Labels["caronte.scale.maxPreplicasPerNode"])

	scaleUpThreshold := labelStringToFloat(annotations.Labels["caronte.metric.scaleUpThreshold"])
	scaleDownThreshold := labelStringToFloat(annotations.Labels["caronte.metric.scaleDownThreshold"])
	store := annotations.Labels["caronte.metric.store"]
	address := annotations.Labels["caronte.metric.prometheus.address"]
	query := annotations.Labels["caronte.metric.query"]
	period := labelStringToInt(annotations.Labels["caronte.metric.aws.period"])
	queue := annotations.Labels["caronte.metric.sqs.queue"]

	provider := annotations.Labels["caronte.instance.provider"]
	filters := annotations.Labels["caronte.instance.aws.asg.filters"]

	caronteService := CaronteService{
		Id:                 id,
		Name:               name,
		ServiceScheduler:   serviceScheduler,
		Max:                max,
		Min:                min,
		MaxReplicasPerNode: maxReplicasPerNode,
		Step:               step,
		ScaleUpThreshold:   scaleUpThreshold,
		ScaleDownThreshold: scaleDownThreshold,
		CoolDown:           coolDown,
		MetricSpecs: metricstores.MetricSpecs{
			Store: store,
			Query: query,
			PrometheusStore: metricstores.MetricPrometheusStore{
				Address: address,
			},
			AwsStore: metricstores.MetricCloudWatchStore{
				Period: period,
			},
			SQSStore: metricstores.MetricSQSStore{
				QueueName: queue,
			},
		},
		InstanceSpecs: instances.ScaleSpecs{
			Provider: provider,
			Aws: instances.AwsScale{
				Filters: filters,
			},
		},
	}

	caronteService.MetricProvider, _ = metricstores.MetricProviderStore{}.GetProvider(caronteService.MetricSpecs)
	caronteService.InstanceProvider, _ = instances.InstanceProviderManager{}.GetProvider(caronteService.InstanceSpecs)

	return caronteService
}

func labelStringToFloat(labelValue string) float64 {
	if labelValue != "" {
		f, err := strconv.ParseFloat(labelValue, 64)
		if err != nil {
			zap.S().Warnf("Fail parsing label value to float %s", labelValue)
			return 0
		}
		return f
	}
	return 0
}

func labelStringToInt(labelValue string) int {
	if labelValue != "" {
		i, err := strconv.Atoi(labelValue)
		if err != nil {
			zap.S().Warnf("Fail parsing label value to float %s", labelValue)
			return 0
		}
		return i
	}
	return 0
}
