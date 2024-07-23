package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = atLeastOneChild{}

// atLeastOneChild validates if the provided object has at least one child attribute.
type atLeastOneChild struct {
}

func (validator atLeastOneChild) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Only validate the attribute configuration value if it is known.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	defined := make(map[string]bool)
	count := 0
	for key, attr := range req.ConfigValue.Attributes() {
		if attr.IsNull() {
			continue
		}
		defined[key] = true
		count++
	}
	if count < 1 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeTypeDiagnostic(
			req.Path,
			validator.MarkdownDescription(ctx),
			req.ConfigValue.String(),
		))
	}
}

func (validator atLeastOneChild) Description(ctx context.Context) string {
	return "value must have at least one child attribute defined"
}

func (validator atLeastOneChild) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

// AtLeastOneChild returns an AttributeValidator which ensures that any configured
// attribute object has at least one child attribute.
// Null (unconfigured) and unknown values are skipped.
func AtLeastOneChild() validator.Object {
	return atLeastOneChild{}
}
