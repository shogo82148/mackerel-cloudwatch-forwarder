package forwarder

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestToMetricDataQuery(t *testing.T) {
	testcases := []struct {
		in  []*Query
		out []types.MetricDataQuery
	}{
		{
			in: []*Query{
				// service metric
				{
					Service: "foo-bar",
					Name:    "metric.sum",
					Metric:  []interface{}{"Namespace", "MetricName"},
					Stat:    "Sum",
				},
				// shorthand
				{
					Service: ".",
					Name:    "metric.average",
					Metric:  []interface{}{".", "."},
					Stat:    "Average",
				},
			},
			out: []types.MetricDataQuery{
				{
					Id:    aws.String("m1"),
					Label: aws.String("service=foo-bar:metric.sum"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("Namespace"),
							MetricName: aws.String("MetricName"),
						},
						Period: aws.Int32(60),
						Stat:   aws.String("Sum"),
					},
				},
				{
					Id:    aws.String("m2"),
					Label: aws.String("service=foo-bar:metric.average"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("Namespace"),
							MetricName: aws.String("MetricName"),
						},
						Period: aws.Int32(60),
						Stat:   aws.String("Average"),
					},
				},
			},
		},

		{
			in: []*Query{
				// host metric
				{
					Host:   "host-foo-bar",
					Name:   "metric.sum",
					Metric: []interface{}{"Namespace", "MetricName", "Host-Dimension1", "foo", "Host-Dimension2", "bar"},
					Stat:   "Sum",
				},
				// shorthand
				{
					Host:   "host-hoge-fuga",
					Name:   "metric.sum",
					Metric: []interface{}{".", ".", ".", "hoge", ".", "fuga"},
					Stat:   "Sum",
				},
			},
			out: []types.MetricDataQuery{
				{
					Id:    aws.String("m1"),
					Label: aws.String("host=host-foo-bar:metric.sum"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("Namespace"),
							MetricName: aws.String("MetricName"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("Host-Dimension1"),
									Value: aws.String("foo"),
								},
								{
									Name:  aws.String("Host-Dimension2"),
									Value: aws.String("bar"),
								},
							},
						},
						Period: aws.Int32(60),
						Stat:   aws.String("Sum"),
					},
				},
				{
					Id:    aws.String("m2"),
					Label: aws.String("host=host-hoge-fuga:metric.sum"),
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							Namespace:  aws.String("Namespace"),
							MetricName: aws.String("MetricName"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("Host-Dimension1"),
									Value: aws.String("hoge"),
								},
								{
									Name:  aws.String("Host-Dimension2"),
									Value: aws.String("fuga"),
								},
							},
						},
						Period: aws.Int32(60),
						Stat:   aws.String("Sum"),
					},
				},
			},
		},
	}

	opt := cmpopts.IgnoreUnexported(types.MetricDataQuery{}, types.MetricStat{}, types.Metric{}, types.Dimension{})
	for _, tc := range testcases {
		got, _, err := ToMetricDataQuery(tc.in)
		if err != nil {
			t.Error(err)
			continue
		}
		if diff := cmp.Diff(got, tc.out, opt); diff != "" {
			t.Errorf("unexpected metric data (-want +got):\n%s", diff)
		}
	}
}
