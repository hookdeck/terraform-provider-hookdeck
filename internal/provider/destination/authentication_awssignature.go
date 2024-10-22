package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type awsSignatureAuthenticationMethodModel struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Region          types.String `tfsdk:"region"`
	Service         types.String `tfsdk:"service"`
}

type awsSignatureAuthenticationMethod struct {
}

func (*awsSignatureAuthenticationMethod) name() string {
	return "aws_signature"
}

func (*awsSignatureAuthenticationMethod) schema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"access_key_id": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `AWS access key id`,
			},
			"secret_access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `AWS secret access key`,
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Sensitive:   false,
				Description: `AWS region`,
			},
			"service": schema.StringAttribute{
				Optional:    true,
				Sensitive:   false,
				Description: `AWS service`,
			},
		},
		Description: `AWS Signature`,
	}
}

func awsSignatureAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"access_key_id":     types.StringType,
		"secret_access_key": types.StringType,
		"region":            types.StringType,
		"service":           types.StringType,
	}
}

func (awsSignatureAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: awsSignatureAuthenticationMethodAttrTypesMap()}
}

func (awsSignatureAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(awsSignatureAuthenticationMethodAttrTypesMap())
}

func (awsSignatureAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.AwsSignature == nil {
		return
	}

	m.AuthMethod.AWSSignature = &awsSignatureAuthenticationMethodModel{}
	m.AuthMethod.AWSSignature.AccessKeyID = types.StringValue(destination.AuthMethod.AwsSignature.Config.AccessKeyId)
	m.AuthMethod.AWSSignature.SecretAccessKey = types.StringValue(destination.AuthMethod.AwsSignature.Config.SecretAccessKey)
	if destination.AuthMethod.AwsSignature.Config.Region != nil {
		m.AuthMethod.AWSSignature.Region = types.StringValue(*destination.AuthMethod.AwsSignature.Config.Region)
	}
	if destination.AuthMethod.AwsSignature.Config.Service != nil {
		m.AuthMethod.AWSSignature.Service = types.StringValue(*destination.AuthMethod.AwsSignature.Config.Service)
	}
}

func (awsSignatureAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.AWSSignature == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromAwsSignature(&hookdeck.AuthAwsSignature{
		Config: &hookdeck.DestinationAuthMethodAwsSignatureConfig{
			AccessKeyId:     method.AWSSignature.AccessKeyID.ValueString(),
			SecretAccessKey: method.AWSSignature.SecretAccessKey.ValueString(),
			Region:          method.AWSSignature.Region.ValueStringPointer(),
			Service:         method.AWSSignature.Service.ValueStringPointer(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &awsSignatureAuthenticationMethod{})
}
