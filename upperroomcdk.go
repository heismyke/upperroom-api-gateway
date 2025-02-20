package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type UpperroomcdkStackProps struct {
	awscdk.StackProps
}

func NewUpperroomcdkStack(scope constructs.Construct, id string, props *UpperroomcdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
  table := awsdynamodb.NewTable(stack, jsii.String("myAuthTablee"), &awsdynamodb.TableProps{
    PartitionKey:  &awsdynamodb.Attribute{
			Name: jsii.String("access_token"),
			Type: awsdynamodb.AttributeType_STRING,
		},
    TableName: jsii.String("AuthTablee"), 
  })

  myFunction := awslambda.NewFunction(stack, jsii.String("upperRoomFunc"), &awslambda.FunctionProps{
    Runtime: awslambda.Runtime_PROVIDED_AL2023(),
    Code : awslambda.AssetCode_FromAsset(jsii.String("lambda/function.zip"), nil),
    Handler: jsii.String("main"),
    Environment: &map[string]*string{
      "CLIENT_ID": jsii.String("622368337104300"),
      "CLIENT_SECRET": jsii.String("7e18989204c77f6731e26713062c852c579"),
      "REDIRECT_URI": jsii.String("https://07hvi2wzr7.execute-api.eu-north-1.amazonaws.com/prod/callback"),
    },
  })
table.GrantReadWriteData(myFunction)

  

api := awsapigateway.NewRestApi(stack, jsii.String("myRestApi"), &awsapigateway.RestApiProps{
    DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
        AllowHeaders: jsii.Strings("Content-Type", "Authorization"), // Allow relevant headers
        AllowMethods: jsii.Strings("GET", "POST", "PUT", "DELETE"),
        AllowOrigins: jsii.Strings("*"), // Adjust for security
    },
    DeployOptions: &awsapigateway.StageOptions{
        LoggingLevel: awsapigateway.MethodLoggingLevel_INFO,
        DataTraceEnabled: jsii.Bool(true), // Log request/response data
        MetricsEnabled: jsii.Bool(true),  // Enable API Gateway metrics
    },
})


callbackResource := api.Root().AddResource(jsii.String("callback"), nil)

integration := awsapigateway.NewLambdaIntegration(myFunction, nil)
callbackResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
    RequestParameters: &map[string]*bool{
        "method.request.querystring.code": jsii.Bool(true), // Enforces "code" is present
    },
    RequestValidatorOptions: &awsapigateway.RequestValidatorOptions{
        ValidateRequestParameters: jsii.Bool(true), // Enable validation
    },
})


return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewUpperroomcdkStack(app, "UpperroomcdkStack", &UpperroomcdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
