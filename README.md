# mackerel-cloudwatch-forwarder

Forward metrics of AWS CloudWatch to Mackerel

> [!WARNING]
> This software is under heavy development and considered ALPHA quality until the version hits v1.0.0.
> Things might be broken, not all features have been implemented, and APIs are likely to change. YOU HAVE BEEN WARNED.

## Prerequisites

- AWS CLI
- AWS Account
- Mackerel API Key

## How to use

mackerel-cloudwatch-forwarder is a Lambda function that forwards AWS CloudWatch metrics to a server.  
In this explanation, we will deploy the Lambda using CloudFormation.  

The following steps are provided and detailed explanations are given in order:

1. Save the secret (Mackerel X-Api-Key) in the AWS Systems Manager Parameter Store
2. Create a CloudFormation template
3. Deploy CloudFormation template

### Save the secret (Mackerel X-Api-Key)

mackerel-cloudwatch-forwarder requires an API key to make use of Mackerel's API. Below is an example of storing this API key as a SecureString in the AWS Systems Manager Parameter Store. Please remember to specify the parameter name in CloudFormation, although the parameter name for storing the API key can be of your choice.

```shell
aws ssm put-parameter --overwrite --name "/api-keys/api.mackerelio.com/headers/X-Api-Key" --value "${MACKEREL_API_KEY}" --type "SecureString"
```

### Create a CloudFormation template
Example:

```yaml
  MetricsForwarder:
    Type: AWS::Serverless::Application
    Properties:
      Location:
        ApplicationId: arn:aws:serverlessrepo:us-east-1:445285296882:applications/mackerel-cloudwatch-forwarder
        SemanticVersion: 0.0.15
      Parameters:
        ParameterName: "/api-keys/api.mackerelio.com/headers/X-Api-Key"
        ForwardSettings: |
            [
              {
                  {
                    "service": "your service name on Mackerel",
                    "name": "metric name on Mackerel",
                    "metric": [ "Namespace", "MetricName", "Dimension1Name", "Dimension1Value",   {} ],
                    "stat": "Sum"
                  },
                  {
                    "hostId": "host id",
                    "name": "metric name on Mackerel",
                    "metric": [ "Namespace", "MetricName", "Dimension1Name", "Dimension1Value",   {} ],
                    "stat": "Sum"
                  }
              }
            ]   
```

The ForwardSettings parameter is expressed in JSON array.  
  
The "metric" key within the JSON is equivalent to the "metrics" key in the JSON displayed on the [Source] tab of AWS CloudWatch Metrics (Path: [CloudWatch] > [Metrics] > [${Your Custom Metrics Name}] > [Source]).

## LICENSE

[MIT LICENCE](./LICENSE)

## Reference
- [サーバーレスでCloudWatchメトリクスをMackerelに転送する (in Japanese)](https://shogo82148.github.io/blog/2019/01/31/mackerel-cloudwatch-transfer/)
