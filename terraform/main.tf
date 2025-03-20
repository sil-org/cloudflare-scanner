
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
        Action = [
          "cloudformation:List*",
          "cloudformation:Get*",
          "cloudformation:ValidateTemplate",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "cloudformation:CreateStack",
          "cloudformation:CreateUploadBucket",
          "cloudformation:DeleteStack",
          "cloudformation:Describe*",
          "cloudformation:UpdateStack",
          "cloudformation:DescribeChangeSet",
          "cloudformation:DeleteChangeSet",
          "cloudformation:CreateChangeSet",
          "cloudformation:ExecuteChangeSet",
        ]
        Effect = "Allow"
        Resource = [
          "arn:aws:cloudformation:${var.aws_region}:*:stack/${var.app_name}",
          "arn:aws:cloudformation:${var.aws_region}:*:stack/aws-sam-cli-managed-default/*",
          "arn:aws:cloudformation:${var.aws_region}:aws:transform/Serverless-2016-10-31",
        ]
      },
      {
        Action = [
          "events:Put*",
          "events:Remove*",
          "events:Delete*",
          "events:DescribeRule"
        ]
        Effect = "Allow"
        Resource = [
          "arn:aws:events:${var.aws_region}::event-source/*",
          "arn:aws:events:${var.aws_region}:*:rule/${var.app_name}*",
          "arn:aws:events:${var.aws_region}:*:event-bus/*",
        ]
      },
      {
        Action = [
          "lambda:Get*",
          "lambda:List*",
          "lambda:CreateFunction",
        ],
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "lambda:AddPermission",
          "lambda:CreateAlias",
          "lambda:DeleteFunction",
          "lambda:InvokeFunction",
          "lambda:PublishVersion",
          "lambda:RemovePermission",
          "lambda:TagResource",
          "lambda:UntagResource",
          "lambda:Update*",
          "lambda:PutFunctionEventInvokeConfig",
          "lambda:DeleteFunctionEventInvokeConfig",
        ]
        Effect = "Allow"
        Resource = [
          "arn:aws:lambda:${var.aws_region}:*:function:${var.app_name}-alert",
          "arn:aws:lambda:${var.aws_region}:*:event-source-mapping:*",
        ]
      },
      {
        Action = [
          "iam:AttachRolePolicy",
          "iam:GetRole",
          "iam:CreateRole",
          "iam:PutRolePolicy",
          "iam:DeleteRolePolicy",
          "iam:DetachRolePolicy",
          "iam:DeleteRole",
          "iam:PassRole",
          "iam:TagRole",
        ]
        Effect   = "Allow"
        Resource = "arn:aws:iam::*:role/${var.app_name}-AlertFunctionRole-*"
      },
      {
        Action = [
          "logs:DeleteLogGroup"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:logs:${var.aws_region}:*:log-group:/aws/lambda/${var.app_name}*"
      },
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:DescribeLogGroups",
          "logs:DescribeLogStreams",
          "logs:PutRetentionPolicy",
          "logs:ListTagsForResource",
          "logs:TagResource",
          "logs:UntagResource"
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "s3:GetBucketLocation",
          "s3:CreateBucket",
          "s3:DeleteBucket",
          "s3:ListBucket",
          "s3:ListBucketVersions",
          "s3:PutBucketAcl",
          "s3:PutAccelerateConfiguration",
          "s3:GetEncryptionConfiguration",
          "s3:PutEncryptionConfiguration",
          "s3:GetBucketPolicy",
          "s3:PutBucketPolicy",
          "s3:DeleteBucketPolicy",
          "s3:PutBucketPublicAccessBlock",
          "s3:PutBucketTagging"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:s3:::${var.app_name}*"
      },
      {
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:s3:::aws-sam-cli-managed-default-samclisourcebucket-*/${var.app_name}/*"
      }
    ],
  })
}
