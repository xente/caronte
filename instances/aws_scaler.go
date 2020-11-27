package instances

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"go.uber.org/zap"
)

var asg *autoscaling.AutoScaling
var autoScalingSession sync.Once

func (a AwsScale) Scale(scaleSpecs ScaleSpecs, direction int) bool {

	autoScalingSession.Do(func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		asg = autoscaling.New(sess)
	})

	autoscalingGroup, err := a.getAsgByTags(scaleSpecs)
	if err != nil {
		zap.S().Error(err)
		return false
	}

	//Only one scaling Group is allowed, it is needed that the filter returns only one
	if len(autoscalingGroup.Tags) == 1 {
		asg, err := a.getAsgByName(autoscalingGroup.Tags[0].ResourceId)
		if err != nil {
			zap.S().Error(err)
			return false
		}

		if len(asg.AutoScalingGroups) == 1 {

			targetAsg := *asg.AutoScalingGroups[0]
			currentCapacity := *targetAsg.DesiredCapacity
			desiredCapacity := *targetAsg.DesiredCapacity + int64(direction*1)

			if a.pendingActivityTasks(targetAsg.AutoScalingGroupName) {
				return false
			}

			if desiredCapacity <= *targetAsg.MaxSize && desiredCapacity >= *targetAsg.MinSize {
				_, err := a.SetDesiredCapacity(targetAsg.AutoScalingGroupName, desiredCapacity)

				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						case autoscaling.ErrCodeScalingActivityInProgressFault:
							return false
						case autoscaling.ErrCodeResourceContentionFault:
							return false
						default:
							zap.S().Error(err)
							return false
						}
					}
				}
				zap.S().Infof("Scale %s from DesiredCapacity=%d to DesiredCapacity=%d", *targetAsg.AutoScalingGroupName, uint(currentCapacity), uint(desiredCapacity))
				return true
			}
		}
	}
	return false
}

func (a AwsScale) RunningInstances(scaleSpecs ScaleSpecs) int {

	autoScalingSession.Do(func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		asg = autoscaling.New(sess)
	})

	autoscalingGroup, err := a.getAsgByTags(scaleSpecs)
	if err != nil {
		zap.S().Error(err)
	}

	if autoscalingGroup.Tags != nil && len(autoscalingGroup.Tags) == 1 {

		asg, _ := a.getAsgByName(autoscalingGroup.Tags[0].ResourceId)
		if len(asg.AutoScalingGroups) == 1 {
			targetAsg := *asg.AutoScalingGroups[0]
			return len(targetAsg.Instances)
		}
	}

	return 0
}

func (a AwsScale) pendingActivityTasks(name *string) bool {
	//https://docs.aws.amazon.com/autoscaling/ec2/userguide/AutoScalingGroupLifecycle.html
	currentActivities, err := a.getAsgActivities(name)
	if err != nil {
		zap.S().Error(err)
		return true
	}

	for _, activity := range currentActivities.Activities {
		if autoscaling.ScalingActivityStatusCodeFailed != *activity.StatusCode &&
			autoscaling.ScalingActivityStatusCodeSuccessful != *activity.StatusCode &&
			autoscaling.ScalingActivityStatusCodeCancelled != *activity.StatusCode &&
			autoscaling.ScalingActivityStatusCodeWaitingForElbconnectionDraining != *activity.StatusCode {
			//zap.S().Debugf("There are pending tasks into asg:  %s", *name)
			return true
		}
	}

	return false
}

func (a AwsScale) PendingTasks(asg *autoscaling.DescribeAutoScalingGroupsOutput) bool {
	for _, asg := range asg.AutoScalingGroups {
		for _, instance := range asg.Instances {
			zap.S().Debugf("Checking Instance %s -  with status %s", *instance.InstanceId, *instance.LifecycleState)

			if *instance.LifecycleState == autoscaling.LifecycleStatePending ||
				*instance.LifecycleState == autoscaling.LifecycleStatePendingProceed ||
				*instance.LifecycleState == autoscaling.LifecycleStatePendingWait {
				return true
			}
		}
	}
	return false
}

func (a AwsScale) SetDesiredCapacity(name *string, capacity int64) (*autoscaling.SetDesiredCapacityOutput, error) {

	input := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(*name),
		DesiredCapacity:      aws.Int64(capacity),
		HonorCooldown:        aws.Bool(true),
	}

	result, err := asg.SetDesiredCapacity(input)

	return result, err
}

func (a AwsScale) getAsgByTags(scaleSpecs ScaleSpecs) (*autoscaling.DescribeTagsOutput, error) {

	if scaleSpecs.Aws.Filters == "" {
		return nil, errors.New("missing aws filters values")
	}

	var filters []*autoscaling.Filter
	err := json.Unmarshal([]byte(scaleSpecs.Aws.Filters), &filters)

	if err != nil {
		return nil, err
	}

	input := &autoscaling.DescribeTagsInput{
		Filters: filters,
	}

	result, err := asg.DescribeTags(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a AwsScale) getAsgActivities(name *string) (*autoscaling.DescribeScalingActivitiesOutput, error) {

	input := &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(*name),
	}

	result, err := asg.DescribeScalingActivities(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a AwsScale) getAsgByName(name *string) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(*name),
		},
	}

	result, err := asg.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}
