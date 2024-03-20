# cloudflare-scanner

Look through Cloudflare records to find the ones that contain a certain substring in their name and
then send emails with that list via AWS SES.

## Credential Rotation

### AWS Serverless User

1. Use the Terraform CLI to taint the old access key
2. Run a new plan on Terraform Cloud
3. Review the new plan and apply if it is correct
4. Copy the new key and secret from the Terraform output into Github Repository Secrets, overwriting the old values
5. Manually rerun the most recent workflow run on the main branch

### Cloudflare

(TBD)
