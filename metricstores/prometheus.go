package metricstores

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

var client api.Client
var v1api v1.API
var prometheusClient sync.Once

type PrometheusStore interface {
	Query(specs MetricSpecs) (float64, error)
}

type MetricPrometheusStore struct {
	Address string
}

func (p MetricPrometheusStore) Query(specs MetricSpecs) (float64, error) {

	prometheusClient.Do(func() {
		cli, err := api.NewClient(api.Config{Address: specs.PrometheusStore.Address})
		if err != nil {
			zap.S().Error(err)
		}
		client = cli
		v1api = v1.NewAPI(client)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, specs.Query, time.Now())

	if err != nil {
		return 0, err
	}

	if len(warnings) > 0 {
		zap.S().Warn(warnings)
	}

	vector := result.(model.Vector)
	if len(vector) == 1 {
		return float64(vector[0].Value), nil
	}

	zap.S().Debugf("prometheus query return invalid data")
	zap.S().Debugf("Query result %g", vector)

	return 0, nil
}
