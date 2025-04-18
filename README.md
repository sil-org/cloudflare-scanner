# cloudflare-scanner

Look through Cloudflare records to find the ones that contain a certain substring in their name and
then send emails with that list via AWS SES.

## AWS SAM

This project uses SAM to deploy to AWS. For development, [install the SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).

## AWS CDK

To build and deploy:

* Build the Go binary:

```sh
CGO_ENABLED=0 go build -C src -tags lambda.norpc -ldflags="-s -w" -o bin/bootstrap ./main.go
```

* Deploy using CDK:

```sh
docker compose run --rm cdk cdk deploy
```

## Credential Rotation

### AWS Serverless User

1. Use the Terraform CLI to taint the old access key
2. Run a new plan on Terraform Cloud
3. Review the new plan and apply if it is correct
4. Copy the new key and secret from the Terraform output into Github Repository Secrets, overwriting the old values
5. Manually rerun the most recent workflow run on the main branch

### Cloudflare

(TBD)
