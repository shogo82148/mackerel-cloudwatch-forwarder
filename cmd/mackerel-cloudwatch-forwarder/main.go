package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	forwarder "github.com/shogo82148/mackerel-cloudwatch-forwarder"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	s := os.Getenv("FORWARD_LOG_LEVEL")
	if s != "" {
		level, err := logrus.ParseLevel(s)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"input": level,
				"error": err,
			}).Error("fail to parse log level")
		} else {
			logrus.SetLevel(level)
		}
	}
}

func main() {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		logrus.WithError(err).Error("fail to load aws config")
	}
	f := &forwarder.Forwarder{
		APIURL: os.Getenv("MACKEREL_APIURL"),
		Config: cfg,
	}
	lambda.Start(f.ForwardMetrics)
}
