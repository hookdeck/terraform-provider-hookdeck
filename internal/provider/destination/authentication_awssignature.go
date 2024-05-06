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
}

type awsSignatureAuthenticationMethod struct {
}

func (*awsSignatureAuthenticationMethod) name() string {
	return "aws_signature"
}

func (*awsSignatureAuthenticationMethod) schema() schema.SingleNestedAttribute {
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
		},
		Description: `AWS Signature`,
	}
}

func (awsSignatureAuthenticationMethod) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"access_key_id":     types.StringType,
		"secret_access_key": types.StringType,
	}
}

func (awsSignatureAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(awsSignatureAuthenticationMethod{}.attrTypes())
}

func (awsSignatureAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.AwsSignature == nil {
		return
	}

	m.AuthMethod.AWSSignature = &awsSignatureAuthenticationMethodModel{}
	m.AuthMethod.AWSSignature.AccessKeyID = types.StringValue(destination.AuthMethod.AwsSignature.Config.AccessKeyId)
	m.AuthMethod.AWSSignature.SecretAccessKey = types.StringValue(destination.AuthMethod.AwsSignature.Config.SecretAccessKey)
}

func (awsSignatureAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.AWSSignature == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromAwsSignature(&hookdeck.AuthAwsSignature{
		Config: &hookdeck.DestinationAuthMethodAwsSignatureConfig{
			AccessKeyId:     method.AWSSignature.AccessKeyID.ValueString(),
			SecretAccessKey: method.AWSSignature.SecretAccessKey.ValueString(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &awsSignatureAuthenticationMethod{})
}
