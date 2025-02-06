package validators

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = HookdeckAPIBaseURLValidator{}

type HookdeckAPIBaseURLValidator struct{}

func (v HookdeckAPIBaseURLValidator) Description(ctx context.Context) string {
	return "value must be a valid base URL without a path (e.g., 'https://api.hookdeck.com')"
}

func (v HookdeckAPIBaseURLValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid base URL without a path (e.g., `https://api.hookdeck.com`)"
}

func (v HookdeckAPIBaseURLValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// If no scheme is provided, try parsing with https:// prefix
	parsedURL, err := url.Parse(value)
	if err != nil || parsedURL.Scheme == "" {
		parsedURL, err = url.Parse("https://" + value)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid URL",
			fmt.Sprintf("Value must be a valid URL: %s", err.Error()),
		)
		return
	}

	if parsedURL.Path != "" && parsedURL.Path != "/" {
		resp.Diagnostics.AddError(
			"Invalid Base URL",
			fmt.Sprintf("URL must not contain a path. Got: %s", value),
		)
		return
	}

	if parsedURL.Host == "" {
		resp.Diagnostics.AddError(
			"Invalid Base URL",
			fmt.Sprintf("URL must include a valid host. Got: %s", value),
		)
		return
	}
}

func NewHookdeckAPIBaseURLValidator() HookdeckAPIBaseURLValidator {
	return HookdeckAPIBaseURLValidator{}
}
