# my-sample-pipeline
Creating sample pipeline for testing codebuild

aws cloudformation create-stack \
  --stack-name CodebuildStack \
  --template-body file://pipeline.yaml \
  --profile aqfer-dev-sso --region us-west-2 \
  --parameters \
    ParameterKey=ProjectName,ParameterValue=TriggerBuild \
    ParameterKey=GitHubOwner,ParameterValue=kalaiselvan45 \
    ParameterKey=GitHubRepo,ParameterValue=my-sample-pipeline \
    ParameterKey=GitHubBranch,ParameterValue=main \
    ParameterKey=GitHubTokenParameterName,ParameterValue=codeBuildToken \
  --capabilities CAPABILITY_IAM

aws cloudformation deploy \
  --template-file pipeline.yaml \
  --stack-name sampleApp \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-west-2 --profile aqfer-dev-sso\
  --parameter-overrides \
    ProjectName=my-sample-pipeline \
    GitHubOwner=Kalaiselvan45 \
    GitHubRepo=my-sample-pipeline \
    GitHubBranch=main \
    GitHubOAuthTokenSecretArn=arn:aws:codeconnections:us-west-2:721176634889:connection/1973e4cf-84bd-4aae-b2cb-55b0573021ff
