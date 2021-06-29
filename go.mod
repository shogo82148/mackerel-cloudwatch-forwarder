module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.16

require (
	github.com/aws/aws-lambda-go v1.24.0
	github.com/aws/aws-sdk-go-v2 v1.7.0
	github.com/aws/aws-sdk-go-v2/config v1.4.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.4.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.3.2
	github.com/aws/aws-sdk-go-v2/service/ssm v1.6.2
	github.com/google/go-cmp v0.5.6
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry v1.1.0
	github.com/sirupsen/logrus v1.8.1
)
