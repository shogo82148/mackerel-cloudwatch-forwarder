module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.17

require (
	github.com/aws/aws-lambda-go v1.34.1
	github.com/aws/aws-sdk-go-v2 v1.16.12
	github.com/aws/aws-sdk-go-v2/config v1.17.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.21.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.18.5
	github.com/aws/aws-sdk-go-v2/service/ssm v1.27.9
	github.com/google/go-cmp v0.5.8
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry v1.1.1
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.12.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.19 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.13 // indirect
	github.com/aws/smithy-go v1.13.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
