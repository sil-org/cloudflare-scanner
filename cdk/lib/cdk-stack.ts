#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib'
import * as lambda from 'aws-cdk-lib/aws-lambda'
import * as iam from 'aws-cdk-lib/aws-iam'
import * as logs from 'aws-cdk-lib/aws-logs'
import * as events from 'aws-cdk-lib/aws-events'
import * as targets from 'aws-cdk-lib/aws-events-targets'
import { Construct } from 'constructs'

export class CdkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    const appID = process.env.APP_ID || ''
    const envID = process.env.ENV_ID || ''
    const configID = process.env.CONFIG_ID || ''
    const functionName = 'cloudflare-scanner'

    const logGroup = new logs.LogGroup(this, `${functionName}-logs`, {
      logGroupName: `/aws/lambda/${functionName}-cdk`,
      retention: logs.RetentionDays.TWO_MONTHS,
      removalPolicy: cdk.RemovalPolicy.DESTROY // Delete logs when stack is deleted
    })

    const lambdaFunction = new lambda.Function(this, `${functionName}-function`, {
      functionName,
      code: lambda.Code.fromAsset('../src/bin'),
      handler: 'bootstrap',
      runtime: lambda.Runtime.PROVIDED_AL2023,
      timeout: cdk.Duration.seconds(300),
      logGroup,
      environment: {
        APP_ID: appID,
        ENV_ID: envID,
        CONFIG_ID: configID,
      },
    })

    const rule = new events.Rule(this, 'ScheduleRule', {
      ruleName: `${functionName}-schedule`,
      schedule: events.Schedule.cron({ minute: '0' }), // Every hour on the hour
    })

    rule.addTarget(new targets.LambdaFunction(lambdaFunction, { retryAttempts: 0 }))

    lambdaFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['ses:SendEmail', 'ses:SendRawEmail'],
        resources: ['*'],
      })
    )

    const appConfigArn = `arn:aws:appconfig:*:*:application/${appID}/environment/${envID}/configuration/${configID}`

    lambdaFunction.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['appconfig:GetLatestConfiguration', 'appconfig:StartConfigurationSession'],
        resources: [appConfigArn],
      })
    )
  }
}

const account = process.env.ACCOUNT_ID
const region = process.env.AWS_REGION

const app = new cdk.App()
new CdkStack(app, 'CdkStack', {
  env: { account, region },
  tags: {
    managed_by: 'cdk',
    itse_app_name: 'cloudflare-scanner',
    itse_app_customer: 'gtis',
    itse_app_env: 'production',
  },
})
