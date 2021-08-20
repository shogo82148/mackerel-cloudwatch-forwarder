module github.com/shogo82148/mackerel-cloudwatch-forwarder

go 1.17

require (
	github.com/aws/aws-lambda-go v1.26.0
	github.com/aws/aws-sdk-go-v2 v1.8.1
	github.com/aws/aws-sdk-go-v2/config v1.6.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.7.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.4.3
	github.com/aws/aws-sdk-go-v2/service/ssm v1.9.0
	github.com/google/go-cmp v0.5.6
	github.com/shogo82148/go-phper-json v0.0.4
	github.com/shogo82148/go-retry v1.1.1
	github.com/sirupsen/logrus v1.8.1
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.6.1 // indirect
	github.com/aws/smithy-go v1.7.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
)
