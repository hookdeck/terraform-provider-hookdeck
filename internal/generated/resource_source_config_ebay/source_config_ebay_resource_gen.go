package resource_source_config_ebay

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func SourceConfigEbayResourceSchema(ctx context.Context) schema.Schema {
	defaultAllowedHttpMethods, _ := types.ListValueFrom(ctx, types.StringType, []string{"POST", "PUT", "PATCH", "DELETE"})

	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_http_methods": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							"GET",
							"POST",
							"PUT",
							"PATCH",
							"DELETE",
						),
					),
				},
				Default: listdefault.StaticValue(defaultAllowedHttpMethods),
			},
			"auth": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Required: true,
					},
					"client_secret": schema.StringAttribute{
						Required: true,
					},
					"dev_id": schema.StringAttribute{
						Required: true,
					},
					"environment": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"PRODUCTION",
								"SANDBOX",
							),
						},
					},
					"verification_token": schema.StringAttribute{
						Required: true,
					},
				},
				CustomType: AuthType{
					ObjectType: types.ObjectType{
						AttrTypes: AuthValue{}.AttributeTypes(ctx),
					},
				},
				Optional: true,
			},
			"source_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type SourceConfigEbayModel struct {
	AllowedHttpMethods types.List   `tfsdk:"allowed_http_methods"`
	Auth               AuthValue    `tfsdk:"auth"`
	SourceId           types.String `tfsdk:"source_id"`
}

var _ basetypes.ObjectTypable = AuthType{}

type AuthType struct {
	basetypes.ObjectType
}

func (t AuthType) Equal(o attr.Type) bool {
	other, ok := o.(AuthType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t AuthType) String() string {
	return "AuthType"
}

func (t AuthType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	clientIdAttribute, ok := attributes["client_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`client_id is missing from object`)

		return nil, diags
	}

	clientIdVal, ok := clientIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`client_id expected to be basetypes.StringValue, was: %T`, clientIdAttribute))
	}

	clientSecretAttribute, ok := attributes["client_secret"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`client_secret is missing from object`)

		return nil, diags
	}

	clientSecretVal, ok := clientSecretAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`client_secret expected to be basetypes.StringValue, was: %T`, clientSecretAttribute))
	}

	devIdAttribute, ok := attributes["dev_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`dev_id is missing from object`)

		return nil, diags
	}

	devIdVal, ok := devIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`dev_id expected to be basetypes.StringValue, was: %T`, devIdAttribute))
	}

	environmentAttribute, ok := attributes["environment"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`environment is missing from object`)

		return nil, diags
	}

	environmentVal, ok := environmentAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`environment expected to be basetypes.StringValue, was: %T`, environmentAttribute))
	}

	verificationTokenAttribute, ok := attributes["verification_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`verification_token is missing from object`)

		return nil, diags
	}

	verificationTokenVal, ok := verificationTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`verification_token expected to be basetypes.StringValue, was: %T`, verificationTokenAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return AuthValue{
		ClientId:          clientIdVal,
		ClientSecret:      clientSecretVal,
		DevId:             devIdVal,
		Environment:       environmentVal,
		VerificationToken: verificationTokenVal,
		state:             attr.ValueStateKnown,
	}, diags
}

func NewAuthValueNull() AuthValue {
	return AuthValue{
		state: attr.ValueStateNull,
	}
}

func NewAuthValueUnknown() AuthValue {
	return AuthValue{
		state: attr.ValueStateUnknown,
	}
}

func NewAuthValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (AuthValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing AuthValue Attribute Value",
				"While creating a AuthValue value, a missing attribute value was detected. "+
					"A AuthValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AuthValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid AuthValue Attribute Type",
				"While creating a AuthValue value, an invalid attribute value was detected. "+
					"A AuthValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AuthValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("AuthValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra AuthValue Attribute Value",
				"While creating a AuthValue value, an extra attribute value was detected. "+
					"A AuthValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra AuthValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewAuthValueUnknown(), diags
	}

	clientIdAttribute, ok := attributes["client_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`client_id is missing from object`)

		return NewAuthValueUnknown(), diags
	}

	clientIdVal, ok := clientIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`client_id expected to be basetypes.StringValue, was: %T`, clientIdAttribute))
	}

	clientSecretAttribute, ok := attributes["client_secret"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`client_secret is missing from object`)

		return NewAuthValueUnknown(), diags
	}

	clientSecretVal, ok := clientSecretAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`client_secret expected to be basetypes.StringValue, was: %T`, clientSecretAttribute))
	}

	devIdAttribute, ok := attributes["dev_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`dev_id is missing from object`)

		return NewAuthValueUnknown(), diags
	}

	devIdVal, ok := devIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`dev_id expected to be basetypes.StringValue, was: %T`, devIdAttribute))
	}

	environmentAttribute, ok := attributes["environment"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`environment is missing from object`)

		return NewAuthValueUnknown(), diags
	}

	environmentVal, ok := environmentAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`environment expected to be basetypes.StringValue, was: %T`, environmentAttribute))
	}

	verificationTokenAttribute, ok := attributes["verification_token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`verification_token is missing from object`)

		return NewAuthValueUnknown(), diags
	}

	verificationTokenVal, ok := verificationTokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`verification_token expected to be basetypes.StringValue, was: %T`, verificationTokenAttribute))
	}

	if diags.HasError() {
		return NewAuthValueUnknown(), diags
	}

	return AuthValue{
		ClientId:          clientIdVal,
		ClientSecret:      clientSecretVal,
		DevId:             devIdVal,
		Environment:       environmentVal,
		VerificationToken: verificationTokenVal,
		state:             attr.ValueStateKnown,
	}, diags
}

func NewAuthValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) AuthValue {
	object, diags := NewAuthValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewAuthValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t AuthType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewAuthValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewAuthValueUnknown(), nil
	}

	if in.IsNull() {
		return NewAuthValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewAuthValueMust(AuthValue{}.AttributeTypes(ctx), attributes), nil
}

func (t AuthType) ValueType(ctx context.Context) attr.Value {
	return AuthValue{}
}

var _ basetypes.ObjectValuable = AuthValue{}

type AuthValue struct {
	ClientId          basetypes.StringValue `tfsdk:"client_id"`
	ClientSecret      basetypes.StringValue `tfsdk:"client_secret"`
	DevId             basetypes.StringValue `tfsdk:"dev_id"`
	Environment       basetypes.StringValue `tfsdk:"environment"`
	VerificationToken basetypes.StringValue `tfsdk:"verification_token"`
	state             attr.ValueState
}

func (v AuthValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 5)

	var val tftypes.Value
	var err error

	attrTypes["client_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["client_secret"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["dev_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["environment"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["verification_token"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 5)

		val, err = v.ClientId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["client_id"] = val

		val, err = v.ClientSecret.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["client_secret"] = val

		val, err = v.DevId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["dev_id"] = val

		val, err = v.Environment.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["environment"] = val

		val, err = v.VerificationToken.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["verification_token"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v AuthValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v AuthValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v AuthValue) String() string {
	return "AuthValue"
}

func (v AuthValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"client_id":          basetypes.StringType{},
		"client_secret":      basetypes.StringType{},
		"dev_id":             basetypes.StringType{},
		"environment":        basetypes.StringType{},
		"verification_token": basetypes.StringType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"client_id":          v.ClientId,
			"client_secret":      v.ClientSecret,
			"dev_id":             v.DevId,
			"environment":        v.Environment,
			"verification_token": v.VerificationToken,
		})

	return objVal, diags
}

func (v AuthValue) Equal(o attr.Value) bool {
	other, ok := o.(AuthValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.ClientId.Equal(other.ClientId) {
		return false
	}

	if !v.ClientSecret.Equal(other.ClientSecret) {
		return false
	}

	if !v.DevId.Equal(other.DevId) {
		return false
	}

	if !v.Environment.Equal(other.Environment) {
		return false
	}

	if !v.VerificationToken.Equal(other.VerificationToken) {
		return false
	}

	return true
}

func (v AuthValue) Type(ctx context.Context) attr.Type {
	return AuthType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v AuthValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":          basetypes.StringType{},
		"client_secret":      basetypes.StringType{},
		"dev_id":             basetypes.StringType{},
		"environment":        basetypes.StringType{},
		"verification_token": basetypes.StringType{},
	}
}
