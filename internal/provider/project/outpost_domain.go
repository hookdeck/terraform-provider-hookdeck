package project

import (
	"context"
	"net/http"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	domainPath = "/config/custom_domain"
)

// customDomain is the portal custom domain as the API returns it.
type customDomain struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	CnameTarget string `json:"cname_target"`
	SSL         struct {
		Status   string `json:"status"`
		Method   string `json:"method"`
		TxtName  string `json:"txt_name"`
		TxtValue string `json:"txt_value"`
	} `json:"ssl"`
	VerificationErrors    []string `json:"verification_errors"`
	OwnershipVerification *struct {
		Type  string `json:"type"`
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"ownership_verification"`
}

var (
	sslTypes = map[string]attr.Type{
		"status":    types.StringType,
		"method":    types.StringType,
		"txt_name":  types.StringType,
		"txt_value": types.StringType,
	}
	ownershipTypes = map[string]attr.Type{
		"type":  types.StringType,
		"name":  types.StringType,
		"value": types.StringType,
	}
	domainStatusTypes = map[string]attr.Type{
		"status":                 types.StringType,
		"cname_target":           types.StringType,
		"ssl":                    types.ObjectType{AttrTypes: sslTypes},
		"ownership_verification": types.ObjectType{AttrTypes: ownershipTypes},
		"verification_errors":    types.ListType{ElemType: types.StringType},
	}
)

// getDomain returns the portal custom domain, or nil when the project has
// none. The API answers 200 with null for no domain and 404 when the project
// has no deployment yet; both mean none.
func getDomain(ctx context.Context, api sdkclient.Client) (*customDomain, error) {
	var domain *customDomain
	err := api.Do(ctx, http.MethodGet, domainPath, nil, nil, &domain)
	if sdkclient.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func addDomain(ctx context.Context, api sdkclient.Client, hostname string) (*customDomain, error) {
	var domain customDomain
	if err := api.Do(ctx, http.MethodPost, domainPath, map[string]string{"hostname": hostname}, nil, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// deleteDomain removes the portal custom domain. A project without a domain,
// or one already deleted, is not an error.
func deleteDomain(ctx context.Context, api sdkclient.Client) error {
	if err := api.Do(ctx, http.MethodDelete, domainPath, nil, nil, nil); err != nil && !isGone(err) {
		return err
	}
	return nil
}

// statusObject maps the API's domain to portal_custom_domain_status.
func (d *customDomain) statusObject(ctx context.Context) (types.Object, diag.Diagnostics) {
	if d == nil {
		return types.ObjectNull(domainStatusTypes), nil
	}
	errs := d.VerificationErrors
	if errs == nil {
		errs = []string{}
	}
	var diags diag.Diagnostics
	errorList, d1 := types.ListValueFrom(ctx, types.StringType, errs)
	ssl, d2 := types.ObjectValue(sslTypes, map[string]attr.Value{
		"status":    types.StringValue(d.SSL.Status),
		"method":    types.StringValue(d.SSL.Method),
		"txt_name":  types.StringValue(d.SSL.TxtName),
		"txt_value": types.StringValue(d.SSL.TxtValue),
	})
	ownership := types.ObjectNull(ownershipTypes)
	var d3 diag.Diagnostics
	if o := d.OwnershipVerification; o != nil {
		ownership, d3 = types.ObjectValue(ownershipTypes, map[string]attr.Value{
			"type":  types.StringValue(o.Type),
			"name":  types.StringValue(o.Name),
			"value": types.StringValue(o.Value),
		})
	}
	diags.Append(d1...)
	diags.Append(d2...)
	diags.Append(d3...)
	if diags.HasError() {
		return types.ObjectNull(domainStatusTypes), diags
	}
	obj, d4 := types.ObjectValue(domainStatusTypes, map[string]attr.Value{
		"status":                 types.StringValue(d.Status),
		"cname_target":           types.StringValue(d.CnameTarget),
		"ssl":                    ssl,
		"ownership_verification": ownership,
		"verification_errors":    errorList,
	})
	diags.Append(d4...)
	return obj, diags
}

// domainValues returns the attribute values for the live domain.
func domainValues(ctx context.Context, live *customDomain) (types.String, types.Object, diag.Diagnostics) {
	status, diags := live.statusObject(ctx)
	if live == nil {
		return types.StringNull(), status, diags
	}
	return types.StringValue(live.Hostname), status, diags
}

func (m *outpostProjectModel) setDomain(ctx context.Context, live *customDomain) (diags diag.Diagnostics) {
	m.PortalCustomDomain, m.PortalCustomDomainStatus, diags = domainValues(ctx, live)
	return diags
}

// domainStatusPlan keeps the status from state while the hostname stays the
// same, and has none when no hostname is set. A new hostname leaves it
// unknown until the domain is added.
type domainStatusPlan struct{}

func (domainStatusPlan) Description(context.Context) string {
	return "Keeps the status from state unless portal_custom_domain changes."
}

func (p domainStatusPlan) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}

func (domainStatusPlan) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	var planned, prior types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("portal_custom_domain"), &planned)...)
	if resp.Diagnostics.HasError() || planned.IsUnknown() {
		return
	}
	if planned.IsNull() {
		resp.PlanValue = types.ObjectNull(domainStatusTypes)
		return
	}
	if req.State.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("portal_custom_domain"), &prior)...)
	if !resp.Diagnostics.HasError() && planned.Equal(prior) && !req.StateValue.IsNull() {
		resp.PlanValue = req.StateValue
	}
}
