package core

import (
	"Caronte/instances"
	"Caronte/metricstores"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"go.uber.org/zap"
)

type CaronteService struct {
	Id                   string
	Name                 string
	ServiceScheduler     int
	Max                  int
	Min                  int
	MaxReplicasPerNode   int
	ServiceCoolDownDelay int
	Step                 int
	ScaleUpThreshold     float64
	ScaleDownThreshold   float64
	Thread               int
	MetricSpecs          metricstores.MetricSpecs
	MetricProvider       metricstores.MetricProvider
	InstanceSpecs        instances.ScaleSpecs
	InstanceProvider     instances.InstanceManagerProvider
	UpdatedAt            time.Time
}

var this CaronteService

func NewCaronteService(id string, name string, annotations swarm.Annotations) CaronteService {

	serviceScheduler := labelStringToInt(annotations.Labels["caronte.ervice.scheduler.scale.time"])
	max := labelStringToInt(annotations.Labels["caronte.scale.max"])
	min := labelStringToInt(annotations.Labels["caronte.scale.min"])
	step := labelStringToInt(annotations.Labels["caronte.scale.step"])
	serviceCoolDownDelay := labelStringToInt(annotations.Labels["caronte.service.coolDownDelay"])
	maxReplicasPerNode := labelStringToInt(annotations.Labels["caronte.scale.maxReplicasPerNode"])

	scaleUpThreshold := labelStringToFloat(annotations.Labels["caronte.metric.scaleUpThreshold"])
	scaleDownThreshold := labelStringToFloat(annotations.Labels["caronte.metric.scaleDownThreshold"])
	store := annotations.Labels["caronte.metric.store"]
	address := annotations.Labels["caronte.metric.prometheus.address"]
	query := annotations.Labels["caronte.metric.query"]
	period := labelStringToInt(annotations.Labels["caronte.metric.aws.period"])
	queue := annotations.Labels["caronte.metric.sqs.queue"]

	provider := annotations.Labels["caronte.instance.provider"]
	instanceCoolDownDelay := labelStringToInt(annotations.Labels["caronte.instance.coolDownDelay"])
	filters := annotations.Labels["caronte.instance.aws.asg.filters"]

	caronteService := CaronteService{
		Id:                   id,
		Name:                 name,
		ServiceScheduler:     serviceScheduler,
		Max:                  max,
		Min:                  min,
		MaxReplicasPerNode:   maxReplicasPerNode,
		ServiceCoolDownDelay: serviceCoolDownDelay,
		Step:                 step,
		ScaleUpThreshold:     scaleUpThreshold,
		ScaleDownThreshold:   scaleDownThreshold,
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
			CoolDown: instanceCoolDownDelay,
			Aws: instances.AwsScale{
				Filters: filters,
			},
		},
	}

	caronteService.MetricProvider, _ = metricstores.MetricProviderStore{}.GetProvider(caronteService.MetricSpecs)
	caronteService.InstanceProvider, _ = instances.InstanceProviderManager{}.GetProvider(caronteService.InstanceSpecs)
	this = caronteService

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
