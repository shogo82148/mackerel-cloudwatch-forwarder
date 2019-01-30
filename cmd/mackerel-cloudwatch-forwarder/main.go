package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
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
	lambda.Start(f.ForwardMetrics)
}
