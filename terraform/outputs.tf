
output "serverless-access-key-id" {
  value = aws_iam_access_key.cdk.id
}

output "serverless-secret-access-key" {
  value     = aws_iam_access_key.cdk.secret
  sensitive = true
}
