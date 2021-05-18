module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.16

require (
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.5.0
	github.com/aws/aws-sdk-go-v2/config v1.2.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.4.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.3.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.6.0
	github.com/google/go-cmp v0.5.5
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/sirupsen/logrus v1.8.1
)
