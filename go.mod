module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.16

require (
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v0.31.0
	github.com/aws/aws-sdk-go-v2/config v0.4.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v0.31.0
	github.com/aws/aws-sdk-go-v2/service/kms v0.31.0
	github.com/aws/aws-sdk-go-v2/service/ssm v0.31.0
	github.com/google/go-cmp v0.5.4
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/sirupsen/logrus v1.8.0
)
