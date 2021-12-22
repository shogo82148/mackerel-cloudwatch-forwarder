module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.17

require (
	github.com/aws/aws-lambda-go v1.27.1
	github.com/aws/aws-sdk-go-v2 v1.11.2
	github.com/aws/aws-sdk-go-v2/config v1.11.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.13.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.11.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.17.1
	github.com/google/go-cmp v0.5.6
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry v1.1.1
	github.com/sirupsen/logrus v1.8.1
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.6.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.11.1 // indirect
	github.com/aws/smithy-go v1.9.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
)
