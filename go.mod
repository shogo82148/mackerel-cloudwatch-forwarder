module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.22.4

require (
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go-v2 v1.35.0
	github.com/aws/aws-sdk-go-v2/config v1.29.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.43.11
	github.com/aws/aws-sdk-go-v2/service/kms v1.37.15
	github.com/aws/aws-sdk-go-v2/service/ssm v1.56.9
	github.com/google/go-cmp v0.6.0
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry v1.3.1
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.17.56 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.11 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
