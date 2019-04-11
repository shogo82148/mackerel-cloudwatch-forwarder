package forwarder

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestToMetricDataQuery(t *testing.T) {
	testcases := []struct {
		in  []*Query
		out []cloudwatch.MetricDataQuery
	}{
		{
			in: []*Query{
				{
					Service: "foo-bar",
					Name:    "some.metric",
					Metric:  []interface{}{"Namespace", "MetricName"},
					Stat:    "Sum",
				},
			},
			out: []cloudwatch.MetricDataQuery{
				{
					Id:    aws.String("m1"),
					Label: aws.String("service=foo-bar:some.metric"),
					MetricStat: &cloudwatch.MetricStat{
						Metric: &cloudwatch.Metric{
							Namespace:  aws.String("Namespace"),
							MetricName: aws.String("MetricName"),
						},
						Period: aws.Int64(60),
						Stat:   aws.String("Sum"),
					},
				},
			},
		},
	}

	opt := cmpopts.IgnoreUnexported(cloudwatch.MetricDataQuery{}, cloudwatch.MetricStat{}, cloudwatch.Metric{})
	for _, tc := range testcases {
		got, err := ToMetricDataQuery(tc.in)
		if err != nil {
			t.Error(err)
			continue
		}
		if diff := cmp.Diff(got, tc.out, opt); diff != "" {
			t.Errorf("unexpected metric data (-want +got):\n%s", diff)
		}
	}
}
