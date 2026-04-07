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

### Cloudflare

(TBD)
