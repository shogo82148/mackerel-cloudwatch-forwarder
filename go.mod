module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.17

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go-v2 v1.22.2
	github.com/aws/aws-sdk-go-v2/config v1.24.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.30.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.26.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.42.1
	github.com/google/go-cmp v0.6.0
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry v1.1.1
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.17.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.25.1 // indirect
	github.com/aws/smithy-go v1.16.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
