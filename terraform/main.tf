resource "aws_iam_user" "cdk" {
  name = "${var.app_name}-cdk"
}

resource "aws_iam_access_key" "cdk" {
  user = aws_iam_user.cdk.name
}

resource "aws_iam_policy" "cdk" {
  name        = "${var.app_name}-cdk"
  description = "CDK deployment policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "sts:AssumeRole"
      Resource = "arn:aws:iam::*:role/cdk-*"
    }]
  })
}
