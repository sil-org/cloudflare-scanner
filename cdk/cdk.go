package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	region := os.Getenv("AWS_REGION")
	appID := os.Getenv("APP_ID")
	envID := os.Getenv("ENV_ID")
	configID := os.Getenv("CONFIG_ID")
	functionName := "CloudflareScanner"

	logGroup := awslogs.NewLogGroup(stack, jsii.String("LambdaLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/" + functionName + "-cdk"),
		Retention:     awslogs.RetentionDays_TWO_MONTHS,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY, // Remove logs when stack is deleted
	})

	function := awslambda.NewFunction(stack, &functionName, &awslambda.FunctionProps{
		Code: awslambda.Code_FromAsset(jsii.String("../src/bin"), nil),
		Environment: &map[string]*string{
			"APP_ID":     &appID,
			"ENV_ID":     &envID,
			"CONFIG_ID":  &configID,
			"SENTRY_DSN": jsii.String(os.Getenv("SENTRY_DSN")),
			"APP_ENV":    jsii.String(getEnv("APP_ENV", "prod")),
		},
		FunctionName:  &functionName,
		Handler:       jsii.String("bootstrap"),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogGroup:      logGroup,
		Runtime:       awslambda.Runtime_PROVIDED_AL2023(),
		Timeout:       awscdk.Duration_Seconds(jsii.Number(300)),
	})

	rule := awsevents.NewRule(stack, jsii.String("ScheduleRule"), &awsevents.RuleProps{
		RuleName: jsii.String(functionName + "-schedule"),
		Schedule: awsevents.Schedule_Cron(&awsevents.CronOptions{
			Minute: jsii.String("0"), // every hour, on the hour
		}),
	})

	rule.AddTarget(awseventstargets.NewLambdaFunction(function, &awseventstargets.LambdaFunctionProps{
		RetryAttempts: jsii.Number(0),
	}))

	function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings(
			"ses:SendEmail",
			"ses:SendRawEmail",
		),
		Resources: jsii.Strings("*"), // Adjust this to restrict access if needed
	}))

	appConfigArn := fmt.Sprintf("arn:aws:appconfig:%s:*:application/%s/environment/%s/configuration/%s",
		region, appID, envID, configID)
	function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings(
			"appconfig:GetLatestConfiguration",
			"appconfig:StartConfigurationSession",
		),
		Resources: jsii.Strings(appConfigArn),
	}))

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkStack(app, "CloudflareScanner", &CdkStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String(os.Getenv("AWS_REGION")),
			},
			Tags: &map[string]*string{
				"managed_by":        jsii.String("cdk"),
				"itse_app_name":     jsii.String("cloudflare-scanner"),
				"itse_app_customer": jsii.String("gtis"),
				"itse_app_env":      jsii.String(getEnv("APP_ENV", "production")),
			},
		},
	})

	app.Synth(nil)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
