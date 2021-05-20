#!/bin/bash

# deploy to the author's (shogo8214's) AWS account for debugging

make all
sam package \
    --template-file template.yaml \
    --output-template-file packaged-test.yaml \
    --s3-bucket shogo82148-test
sam deploy \
    --template-file packaged-test.yaml \
    --stack-name mackerel-cloudwatch-forwarder-test \
    --capabilities CAPABILITY_IAM \
    --parameter-overrides ParameterName=/development/api.mackerelio.com/headers/X-Api-Key
