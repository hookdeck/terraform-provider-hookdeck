package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = atMostOneOf{}

// atMostOneOf validates that at most one of the specified attributes is set.
type atMostOneOf struct {
	attributes []string
}

func (v atMostOneOf) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// Only validate the attribute configuration value if it is known.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var setAttributes []string
	for _, attrName := range v.attributes {
		if attr, exists := req.ConfigValue.Attributes()[attrName]; exists && !attr.IsNull() && !attr.IsUnknown() {
			setAttributes = append(setAttributes, attrName)
		}
	}

	if len(setAttributes) > 1 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("Only one of [%s] can be specified, but %s were specified",
				strings.Join(v.attributes, ", "),
				strings.Join(setAttributes, " and ")),
		))
	}
}

func (v atMostOneOf) Description(ctx context.Context) string {
	return fmt.Sprintf("at most one of %s can be specified", strings.Join(v.attributes, ", "))
}

func (v atMostOneOf) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// AtMostOneOf returns an AttributeValidator which ensures that at most one
// of the specified attributes is set (non-null).
func AtMostOneOf(attributes ...string) validator.Object {
	return atMostOneOf{
		attributes: attributes,
	}
}
