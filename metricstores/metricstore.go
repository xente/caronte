package metricstores

import (
	"errors"
)

type MetricSpecs struct {
	Store           string
	Query           string
	PrometheusStore MetricPrometheusStore
	AwsStore        MetricCloudWatchStore
	SQSStore        MetricSQSStore
}

const (
	Prometheus = "prometheus"
	CloudWatch = "cloudwatch"
	SQS        = "sqs"
)

type MetricProviderStore struct {
}

type MetricStore interface {
	GetProvider(specs MetricSpecs) (MetricProvider, error)
}

type MetricProvider interface {
	Query(specs MetricSpecs) (float64, error)
}

func (m MetricProviderStore) GetProvider(specs MetricSpecs) (MetricProvider, error) {

	switch specs.Store {
	case Prometheus:
		return MetricPrometheusStore{
			specs.PrometheusStore.Address,
		}, nil
	case CloudWatch:
		return MetricCloudWatchStore{
			specs.AwsStore.Period,
		}, nil
	case SQS:
		return MetricSQSStore{}, nil
	}

	return nil, errors.New("metric provided required")
}
