module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.26

require (
	github.com/aws/aws-lambda-go v1.55.0
	github.com/aws/aws-sdk-go-v2 v1.45.1
	github.com/aws/aws-sdk-go-v2/config v1.33.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.69.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.57.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.75.1
	github.com/google/go-cmp v0.7.0
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry/v2 v2.0.2
	github.com/sirupsen/logrus v1.10.2
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.35.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.40.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.47.1 // indirect
	github.com/aws/smithy-go v1.28.1 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
