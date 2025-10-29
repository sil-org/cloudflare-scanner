# cloudflare-scanner

Look through Cloudflare records to find the ones that contain a certain substring in their name and
then send emails with that list via AWS SES.

## AWS CDK

This project uses CDK to deploy to AWS. For development, use Docker Compose or [install the CDK CLI](https://docs.aws.amazon.com/cdk/v2/guide/getting-started.html#getting-started-install).

To build and deploy:

* Build the Go binary:

```sh
CGO_ENABLED=0 go build -C src -tags lambda.norpc -ldflags="-s -w" -o bin/bootstrap ./main.go
```

* Deploy using CDK:

```sh
docker compose run --rm cdk cdk deploy
```

or simply `cdk deploy` if you installed the CLI.

## Credential Rotation

### AWS CDK User

1. Run a new plan on Terraform Cloud
   1. On the Create Run screen click `Additional Planning Options`
   2. Under `Replace Resources` choose `aws_iam_access_key.cdk`
2. Review the new plan and apply if it is correct
3. Copy the new key and secret from the Terraform output into Github Repository Secrets, overwriting the old values
4. Manually rerun the most recent workflow run on the main branch

### Cloudflare

(TBD)
