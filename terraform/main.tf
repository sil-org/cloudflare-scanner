
/*
 * Create IAM user for Serverless framework to use to deploy the lambda function
 */
module "serverless-user" {
  source  = "silinternational/serverless-user/aws"
  version = "~> 0.4.2"

  app_name = var.app_name

  policy_override = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sts:AssumeRole",
        ]
        Resource = [
          "arn:aws:iam::*:role/cdk-*"
        ]
      }
    ],
  })
}
