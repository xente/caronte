package metricstores

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"go.uber.org/zap"
)

var clw *cloudwatch.CloudWatch
var cloudWatchSession sync.Once

type CloudWatchStore interface {
	Query(specs MetricSpecs) (float64, error)
}

type MetricCloudWatchStore struct {
	Period int
}

func (p MetricCloudWatchStore) Query(specs MetricSpecs) (float64, error) {
	cloudWatchSession.Do(func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		clw = cloudwatch.New(sess)
	})

	var queries []*cloudwatch.MetricDataQuery

	queries = append(queries, &cloudwatch.MetricDataQuery{
		Id:         aws.String(fmt.Sprintf("caronte_%x", md5.Sum([]byte(specs.Query)))),
		Expression: aws.String(specs.Query),
		Period:     aws.Int64(int64(specs.AwsStore.Period)),
	})

	result, err := clw.GetMetricData(&cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(time.Now().Add(-time.Duration(specs.AwsStore.Period) * time.Second)),
		EndTime:           aws.Time(time.Now()),
		MetricDataQueries: queries,
	})

	var total float64
	if err != nil {
		zap.S().Debug(err)
	} else if result.MetricDataResults != nil {

		for _, result := range result.MetricDataResults {
			for _, value := range result.Values {
				total += *value
			}
			total = total / float64(len(result.Values))
		}
		return total, nil
	}
	return total, nil
}
