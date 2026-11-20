package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
		Optional:    true,
		Description: "Hostname that serves the project's portal, for example `portal.example.com`. Give it in lowercase: the API stores it as typed, and an uppercase hostname does not match browser requests. Changing it deletes the current domain and adds the new one. `portal_custom_domain_status` has the DNS records to create. A domain added outside Terraform is removed on the next apply unless it is set here.",
	}
	attributes["portal_custom_domain_status"] = schema.SingleNestedAttribute{
		Computed:      true,
		PlanModifiers: []planmodifier.Object{domainStatusPlan{}},
		Description:   domainStatusDescription,
		Attributes: map[string]schema.Attribute{
			"status":       schema.StringAttribute{Computed: true, Description: domainStatusFieldDescription},
			"cname_target": schema.StringAttribute{Computed: true, Description: domainCnameDescription},
			"ssl": schema.SingleNestedAttribute{
				Computed:    true,
				Description: domainSSLDescription,
				Attributes: map[string]schema.Attribute{
					"status":    schema.StringAttribute{Computed: true, Description: domainSSLStatusDescription},
					"method":    schema.StringAttribute{Computed: true, Description: domainSSLMethodDescription},
					"txt_name":  schema.StringAttribute{Computed: true, Description: domainSSLTxtNameDescription},
					"txt_value": schema.StringAttribute{Computed: true, Description: domainSSLTxtValueDescription},
				},
			},
			"ownership_verification": schema.SingleNestedAttribute{
				Computed:    true,
				Description: domainOwnershipDescription,
				Attributes: map[string]schema.Attribute{
					"type":  schema.StringAttribute{Computed: true, Description: "Record type"},
					"name":  schema.StringAttribute{Computed: true, Description: "Record name"},
					"value": schema.StringAttribute{Computed: true, Description: "Record value"},
				},
			},
			"verification_errors": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: domainErrorsDescription},
		},
	}
	return attributes
}

// The status attributes mirror the response of GET /config/custom_domain.
const (
	domainStatusDescription      = "The portal custom domain as the API reports it. Null when no domain is set. Refreshed on every plan, so `status` moves from `pending` to `active` once the DNS records are in place."
	domainStatusFieldDescription = "Domain status: `pending` until the hostname is verified, then `active`"
	domainCnameDescription       = "Host to point a CNAME record for the hostname at"
	domainSSLDescription         = "Certificate issued for the hostname"
	domainSSLStatusDescription   = "Certificate status, for example `initializing`, `pending_validation` or `active`"
	domainSSLMethodDescription   = "Method used to validate the certificate, for example `txt`"
	domainSSLTxtNameDescription  = "Name of the TXT record that validates the certificate. Can be empty right after the domain is added."
	domainSSLTxtValueDescription = "Value of the TXT record that validates the certificate. Can be empty right after the domain is added."
	domainOwnershipDescription   = "TXT record that verifies ownership of the hostname. Null when the API returns none."
	domainErrorsDescription      = "Reasons the domain could not be verified, if any"
)

func outpostTimeoutsBlock(ctx context.Context) schema.Block {
	return timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})
}
