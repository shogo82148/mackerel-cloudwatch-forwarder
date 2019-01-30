package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	forwarder "github.com/shogo82148/mackerel-cloudwatch-forwarder"
)

func main() {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal(err)
	}
	f := &forwarder.Forwarder{
		Config: cfg,
	}
	if err := f.ForwardMetrics(context.Background(), forwarder.ForwardMetricsEvent{
		ServiceMetrics: []forwarder.ServiceMetricDefinition{
			{
				Service: "test-service",
				Name:    "foobar",
				Metric:  []string{"wordpress/shogo82148", "error_count"},
				Stat:    "Sum",
			},
		},
	}); err != nil {
		log.Println(err)
	}
}
