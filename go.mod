module github.com/shogo82148/mackerel-cloudwatch-forwarder

require (
	github.com/aws/aws-lambda-go v1.22.0
	github.com/aws/aws-sdk-go-v2 v0.31.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v0.31.0
	github.com/aws/aws-sdk-go-v2/service/kms v0.31.0
	github.com/aws/aws-sdk-go-v2/service/ssm v0.31.0
	github.com/google/go-cmp v0.5.4
	github.com/shogo82148/go-phper-json v0.0.3
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/sys v0.0.0-20200317113312-5766fd39f98d // indirect
)

go 1.13
