package main

import (
	"fmt"
	"log"
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
	if props == nil {
		panic("props is nil")
	}
	if props.Env == nil {
		panic("props.Env is nil")
	}
	if props.Env.Region == nil {
		panic("props.Env.Region is nil")
	}
	if props.Env.Account == nil {
		panic("props.Env.Account is nil")
	}

	sp := props.StackProps
	stack := awscdk.NewStack(scope, &id, &sp)
	region := *props.Env.Region
	account := *props.Env.Account

	functionName := "CloudflareScanner"

	logGroup := awslogs.NewLogGroup(stack, jsii.String("LambdaLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/" + functionName + "-cdk"),
		Retention:     awslogs.RetentionDays_TWO_MONTHS,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY, // Remove logs when stack is deleted
	})

	function := awslambda.NewFunction(stack, &functionName, &awslambda.FunctionProps{
		Code: awslambda.Code_FromAsset(jsii.String("../src/bin"), nil),
		Environment: &map[string]*string{
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

	parameterArn := fmt.Sprintf("arn:aws:ssm:%s:%s:parameter/cloudflare-scanner/*", region, account)
	function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("ssm:GetParametersByPath"),
		Resources: jsii.Strings(parameterArn),
	}))
	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	account := os.Getenv("AWS_ACCOUNT_ID")
	if account == "" {
		log.Println("AWS_ACCOUNT_ID is not set")
		return
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Println("AWS_REGION is not set")
		return
	}

	NewCdkStack(app, "CloudflareScanner", &CdkStackProps{
		awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String(account),
				Region:  jsii.String(region),
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
