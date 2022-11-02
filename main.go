package main

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	awsprovider "github.com/cdktf/cdktf-provider-aws-go/aws/v10/provider"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/subnet"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/vpc"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

// VPC creates a VPC with a public and private subnet
func newVPC(stack cdktf.TerraformStack, tags *map[string]*string, cidr cdktf.TerraformVariable) vpc.Vpc {
	return vpc.NewVpc(stack, jsii.String("vpc"), &vpc.VpcConfig{
		CidrBlock:          cidr.StringValue(),
		EnableDnsSupport:   jsii.Bool(true),
		EnableDnsHostnames: jsii.Bool(true),
		InstanceTenancy:    jsii.String("default"),
		Tags:               tags,
	})
}

// newPrivateSubnet creates a private subnet
func newPrivateSubnet(stack cdktf.TerraformStack, tags *map[string]*string, cidr, az cdktf.TerraformVariable, vpcID vpc.Vpc) subnet.Subnet {
	return subnet.NewSubnet(stack, jsii.String("private_subnet"), &subnet.SubnetConfig{
		AvailabilityZone: az.StringValue(),
		CidrBlock:        cidr.StringValue(),
		VpcId:            vpcID.Id(),
		Tags:             tags,
	})
}

// newPublicSubnet creates a private subnet
func newPublicSubnet(stack cdktf.TerraformStack, tags *map[string]*string, cidr, az cdktf.TerraformVariable, vpcID vpc.Vpc) subnet.Subnet {
	return subnet.NewSubnet(stack, jsii.String("public_subnet"), &subnet.SubnetConfig{
		AvailabilityZone:    az.StringValue(),
		CidrBlock:           cidr.StringValue(),
		VpcId:               vpcID.Id(),
		MapPublicIpOnLaunch: jsii.Bool(false),
		Tags:                tags,
	})
}

func NewCDKIac(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)
	// Variables
	region := cdktf.NewTerraformVariable(stack, jsii.String("AWS_REGION"), &cdktf.TerraformVariableConfig{
		Default:     "us-east-1",
		Description: jsii.String("Choose which AWS region to use"),
		Nullable:    jsii.Bool(false),
		Sensitive:   jsii.Bool(false),
		Type:        jsii.String("string"),
	})
	tags := &map[string]*string{
		"Project":     jsii.String("cdktf"),
		"Environment": jsii.String("dev"),
		"Name":        jsii.String("cdktf-resources"),
	}
	vpcCidr := cdktf.NewTerraformVariable(stack, jsii.String("VPC_CIDR"), &cdktf.TerraformVariableConfig{
		Type:        jsii.String("string"),
		Description: jsii.String("The CIDR block for the VPC"),
		Default:     jsii.String("10.0.0.0/16"),
	})
	publicCidr := cdktf.NewTerraformVariable(stack, jsii.String("PUBLIC_CIDR"), &cdktf.TerraformVariableConfig{
		Type:        jsii.String("string"),
		Description: jsii.String("The CIDR block for the public subnet"),
		Default:     jsii.String("10.0.1.0/24"),
	})
	privateCidr := cdktf.NewTerraformVariable(stack, jsii.String("PRIVATE_CIDR"), &cdktf.TerraformVariableConfig{
		Type:        jsii.String("string"),
		Description: jsii.String("The CIDR block for the private subnet"),
		Default:     jsii.String("10.0.2.0/24"),
	})

	az := cdktf.NewTerraformVariable(stack, jsii.String("VPC_AZ"), &cdktf.TerraformVariableConfig{
		Type:        jsii.String("string"),
		Description: jsii.String("The availability zone for the VPC"),
		Default:     jsii.String("us-east-1a"),
	})

	awsprovider.NewAwsProvider(stack, jsii.String("AWS"), &awsprovider.AwsProviderConfig{
		Region: region.StringValue(),
	})

	// Create VPC
	VPC := newVPC(stack, tags, vpcCidr)
	// Create public subnet
	PublicSubnet := newPrivateSubnet(stack, tags, publicCidr, az, VPC)
	// Create private subnet
	PrivateSubnet := newPublicSubnet(stack, tags, privateCidr, az, VPC)

	// Stack outputs

	cdktf.NewTerraformOutput(stack, jsii.String("vpc_id"), &cdktf.TerraformOutputConfig{
		Value: VPC.Id(),
	})
	cdktf.NewTerraformOutput(stack, jsii.String("vpc_arn"), &cdktf.TerraformOutputConfig{
		Value: VPC.Arn(),
	})
	cdktf.NewTerraformOutput(stack, jsii.String("vpc_cidr"), &cdktf.TerraformOutputConfig{
		Value: VPC.CidrBlock(),
	})
	cdktf.NewTerraformOutput(stack, jsii.String("private_subnet_id"), &cdktf.TerraformOutputConfig{
		Value: PrivateSubnet.Id(),
	})
	cdktf.NewTerraformOutput(stack, jsii.String("public_subnet_id"), &cdktf.TerraformOutputConfig{
		Value: PublicSubnet.Id(),
	})

	return stack
}
func main() {
	app := cdktf.NewApp(nil)
	stack := NewCDKIac(app, "VPC")
	// S3 bucket for remote state
	cdktf.NewS3Backend(stack, &cdktf.S3BackendProps{
		Bucket:   jsii.String("s3-remote-state-20221018154517921000000001"),
		Key:      jsii.String("cdktf-state/terraform.tfstate"),
		Region:   jsii.String("us-east-1"),
		Encrypt:  jsii.Bool(true),
		KmsKeyId: jsii.String("alias/terraform-bucket-key"),
	})
	app.Synth()
}
