locals {
  aws_account = data.aws_caller_identity.this.account_id
}

data "aws_caller_identity" "this" {}

# Role for Continuous Deployment using CDK

resource "aws_iam_role" "cd" {
  description = "for GitHub Actions to deploy ${var.github_repository}"
  name        = "${var.app_name}-cd"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "GitHub"
      Effect = "Allow"
      Action = "sts:AssumeRoleWithWebIdentity"
      Principal = {
        Federated = var.github_oidc_provider_arn
      }
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" : "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" : "repo:${var.github_repository}:ref:refs/heads/main"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "cd" {
  name = "${var.app_name}-cd"
  role = aws_iam_role.cd.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "sts:AssumeRole"
        Resource = "arn:aws:iam::${local.aws_account}:role/cdk-*"
      },
      {
        Sid      = "SendEmailForActions"
        Effect   = "Allow"
        Action   = "ses:SendEmail"
        Resource = "*"
        Condition = {
          StringEquals = {
            "ses:FromAddress" = "gtis_itse_alerts@groups.sil.org"
          }
          "ForAllValues:StringEquals" = {
            "ses:Recipients" = "gtis_itse_support@sil.org"
          }
        }
      },
    ]
  })
}
