package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func outpostSchemaAttributes() map[string]schema.Attribute {
	attributes := schemaAttributes(outpost)
	attributes["config"] = schema.MapAttribute{
		Optional:    true,
		ElementType: types.StringType,
		Description: "Outpost configuration to manage, keyed by Outpost environment variable name (for example `TOPICS` or `RETRY_INTERVAL_SECONDS`). Only the keys set here are managed and tracked; other keys keep their value. Removing a key restores its default. Keys that can carry credentials belong in `sensitive_config`.",
	}
	attributes["sensitive_config"] = schema.MapAttribute{
		Optional:    true,
		Sensitive:   true,
		ElementType: types.StringType,
		Description: "Outpost configuration to manage like `config`, for keys whose values are secrets or can carry credentials, such as `OTEL_EXPORTER_OTLP_HEADERS`, `DESTINATIONS_PROXY_URL` and the `PUBLISH_*` connection strings and secret keys. Values are stored in the Terraform state.",
	}
	attributes["portal_custom_domain"] = schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.LengthAtMost(253),
			stringvalidator.RegexMatches(hostnamePattern, "must be a lowercase hostname such as portal.example.com, without scheme, port or path"),
		},
		Description: "Hostname that serves the project's portal, for example `portal.example.com`. Changing it deletes the current domain and adds the new one. `portal_custom_domain_status` lists the DNS records to create. A domain added outside Terraform is removed on the next apply unless it is set here.",
	}
	attributes["portal_custom_domain_status"] = schema.SingleNestedAttribute{
		Computed:      true,
		PlanModifiers: []planmodifier.Object{domainStatusPlan{}},
		Description:   domainStatusDescription,
		Attributes: map[string]schema.Attribute{
			"status":     schema.StringAttribute{Computed: true, Description: domainStatusFieldDescription},
			"ssl_status": schema.StringAttribute{Computed: true, Description: domainSSLStatusDescription},
			"verification_records": schema.ListNestedAttribute{
				Computed:    true,
				Description: domainRecordsDescription,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"type":  schema.StringAttribute{Computed: true, Description: "Record type, for example `txt`"},
					"name":  schema.StringAttribute{Computed: true, Description: "Record name"},
					"value": schema.StringAttribute{Computed: true, Description: "Record value"},
				}},
			},
			"verification_errors": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: domainErrorsDescription},
		},
	}
	return attributes
}

const (
	domainStatusDescription      = "State of the portal custom domain. Null when no domain is set. Refreshed on every plan, so it moves from `pending` to `active` once the DNS records are in place."
	domainStatusFieldDescription = "Domain status: `pending` until the hostname is verified, then `active`"
	domainSSLStatusDescription   = "Certificate status, for example `pending_validation` or `active`"
	domainRecordsDescription     = "DNS records to create to verify the domain and issue its certificate"
	domainErrorsDescription      = "Reasons the domain could not be verified, if any"
)

func outpostTimeoutsBlock(ctx context.Context) schema.Block {
	return timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})
}
