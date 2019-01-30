package forwarder

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

func TestParseMetric(t *testing.T) {
	cases := []struct {
		in  interface{}
		out *cloudwatch.GetMetricStatisticsInput
	}{
		{
			// The simplest example, a metric with no dimensions
			in: []interface{}{"AWS/EC2", "CPUUtilization"},
			out: &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("AWS/EC2"),
				MetricName: aws.String("CPUUtilization"),
				Dimensions: []cloudwatch.Dimension{},
				Statistics: []cloudwatch.Statistic{"Average"},
			},
		},
		{
			// A metric with a single dimension
			in: []interface{}{"AWS/EC2", "CPUUtilization", "InstanceId", "i-012345"},
			out: &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("AWS/EC2"),
				MetricName: aws.String("CPUUtilization"),
				Dimensions: []cloudwatch.Dimension{{
					Name:  aws.String("InstanceId"),
					Value: aws.String("i-012345"),
				}},
				Statistics: []cloudwatch.Statistic{"Average"},
			},
		},

		{
			// string JSON format
			in: `["AWS/EC2", "CPUUtilization"]`,
			out: &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("AWS/EC2"),
				MetricName: aws.String("CPUUtilization"),
				Dimensions: []cloudwatch.Dimension{},
				Statistics: []cloudwatch.Statistic{"Average"},
			},
		},
		{
			// bytes JSON format
			in: []byte(`["AWS/EC2", "CPUUtilization"]`),
			out: &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("AWS/EC2"),
				MetricName: aws.String("CPUUtilization"),
				Dimensions: []cloudwatch.Dimension{},
				Statistics: []cloudwatch.Statistic{"Average"},
			},
		},
		{
			// string slice
			in: []string{"AWS/EC2", "CPUUtilization"},
			out: &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("AWS/EC2"),
				MetricName: aws.String("CPUUtilization"),
				Dimensions: []cloudwatch.Dimension{},
				Statistics: []cloudwatch.Statistic{"Average"},
			},
		},
	}

	template := &cloudwatch.GetMetricStatisticsInput{
		Statistics: []cloudwatch.Statistic{"Average"},
	}
	for i, tc := range cases {
		got, err := ParseMetric(template, tc.in)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.out) {
			t.Errorf("%d: want %#v, got %#v", i, tc.out, got)
		}
	}
}
