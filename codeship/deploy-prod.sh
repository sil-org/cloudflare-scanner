#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/build.sh

# Export env vars
export AWS_REGION="${AWS_REGION}"
export CF_API_EMAIL="${CF_API_EMAIL}"
export CF_API_KEY="${CF_API_KEY}"
export CF_CONTAINS_STRING="${CF_CONTAINS_STRING}"
export CF_CONTAINS_STRING_USA="${CF_CONTAINS_STRING_USA}"
export CF_ZONE_NAMES="${CF_ZONE_NAMES}"
export SES_SUBJECT="${SES_SUBJECT}"
export SES_RETURN_TO_ADDR="${SES_RETURN_TO_ADDR}"
export SES_RECIPIENT_EMAILS_USA="${SES_RECIPIENT_EMAILS_USA}"

serverless deploy -v --stage prod