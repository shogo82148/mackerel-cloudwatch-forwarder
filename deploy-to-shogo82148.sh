#!/bin/bash

# deploy to the author's (shogo8214's) AWS account for debugging

make all
sam package \
    --template-file template.yaml \
    --output-template-file packaged-test.yaml \
    --s3-bucket shogo82148-test
ForwardSettings=$(jq -c <<SETTING
[
    {
        "service": "shogo82148",
        "name": "mackerel-cloudwatch-forwarder.duration",
        "stat": "Sum",
        "metric": ["AWS/Lambda", "Duration", "FunctionName", "mackerel-cloudwatch-forwarder-test-Forwarder-oDTgAaRmNR4h"],
        "default": 0
    },
    {
        "service": "shogo82148",
        "name": "grongish.count",
        "stat": "Sum",
        "metric": [ "AWS/ApiGateway", "Count", "ApiName", "Grongish" ],
        "default": 0
    }
]
SETTING
)
sam deploy \
    --template-file packaged-test.yaml \
    --stack-name mackerel-cloudwatch-forwarder-test \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides \
        ParameterName=/development/api.mackerelio.com/headers/X-Api-Key \
        ForwardSettings="'$ForwardSettings'" \
        LogLevel=debug
