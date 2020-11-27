package instances

import (
	"errors"
	"time"
)

type ScaleSpecs struct {
	Provider  string
	CoolDown  int
	UpdatedAt time.Time
	Aws       AwsScale
}

type AwsScale struct {
	Filters string
}

const AWS = "aws"

type InstanceProviderManager struct {
}

type InstanceManagerProvider interface {
	Scale(scaleSpecs ScaleSpecs, direction int) bool
	RunningInstances(scaleSpecs ScaleSpecs) int
}

type InstanceManager interface {
	GetProvider(provider string) (InstanceManagerProvider, error)
}

func (i InstanceProviderManager) GetProvider(specs ScaleSpecs) (InstanceManagerProvider, error) {

	switch specs.Provider {
	case AWS:
		return AwsScale{
			Filters: specs.Aws.Filters,
		}, nil
	}

	return nil, errors.New("metric provided required")
}
