variable "aws_region" {
  description = "A valid AWS region where this lambda will be deployed"
  default     = "us-east-1"
}

variable "aws_access_key" {
  description = "Access Key ID for user with permissions to create the Serverless deployment user"
  default     = null
}

variable "aws_secret_key" {
  description = "Secret access Key ID for user with permissions to create the Serverless deployment user"
  default     = null
}

/*
 * AWS tag values
 */

variable "app_customer" {
  description = "customer name to use for the itse_app_customer tag"
  type        = string
  default     = "gtis"
}

variable "app_environment" {
  description = "environment name to use for the itse_app_environment tag, e.g. staging, production"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "app name to use for the itse_app_name tag"
  type        = string
  default     = "cloudflare-scanner"
}
