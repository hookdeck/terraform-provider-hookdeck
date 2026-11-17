package project

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	domainPath = "/config/custom_domain"

	codeDomainAlreadyConfigured = "CUSTOM_DOMAIN_ALREADY_CONFIGURED"
	codeDomainAlreadyExists     = "CUSTOM_DOMAIN_ALREADY_EXISTS"
	codeDomainInvalid           = "CUSTOM_DOMAIN_INVALID"
)

// customDomain is the portal custom domain as the API returns it.
type customDomain struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	SSL      struct {
		Status   string `json:"status"`
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
	recordTypes = map[string]attr.Type{
		"type":  types.StringType,
		"name":  types.StringType,
		"value": types.StringType,
	}
	domainStatusTypes = map[string]attr.Type{
		"status":               types.StringType,
		"ssl_status":           types.StringType,
		"verification_records": types.ListType{ElemType: types.ObjectType{AttrTypes: recordTypes}},
		"verification_errors":  types.ListType{ElemType: types.StringType},
	}
	hostnamePattern = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+([a-z]{2,63}|xn--[a-z0-9-]{1,59})$`)
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
	err := api.Do(ctx, http.MethodPost, domainPath, map[string]string{"hostname": hostname}, nil, &domain)
	switch {
	case sdkclient.IsCode(err, http.StatusConflict, codeDomainAlreadyConfigured):
		return nil, fmt.Errorf("the project already has a portal custom domain: %w", err)
	case sdkclient.IsCode(err, http.StatusConflict, codeDomainAlreadyExists):
		return nil, fmt.Errorf("%s is already in use by another project: %w", hostname, err)
	// The API answers an invalid hostname with 429.
	case sdkclient.IsCode(err, http.StatusTooManyRequests, codeDomainInvalid):
		return nil, fmt.Errorf("%s is not a valid hostname: %w", hostname, err)
	case err != nil:
		return nil, err
	}
	return &domain, nil
}

// deleteDomain removes the portal custom domain. A project without a domain,
// or one already deleted, is not an error.
func deleteDomain(ctx context.Context, api sdkclient.Client) error {
	err := api.Do(ctx, http.MethodDelete, domainPath, nil, nil, nil)
	var refused *sdkclient.ProjectAccessError
	if sdkclient.IsNotFound(err) || (errors.As(err, &refused) && refused.Status == http.StatusNotFound) {
		return nil
	}
	return err
}

// statusObject maps the API's domain to portal_custom_domain_status.
func (d *customDomain) statusObject(ctx context.Context) (types.Object, diag.Diagnostics) {
	if d == nil {
		return types.ObjectNull(domainStatusTypes), nil
	}
	type record struct {
		Type  string `tfsdk:"type"`
		Name  string `tfsdk:"name"`
		Value string `tfsdk:"value"`
	}
	records := []record{}
	if d.SSL.TxtName != "" {
		records = append(records, record{Type: "txt", Name: d.SSL.TxtName, Value: d.SSL.TxtValue})
	}
	if o := d.OwnershipVerification; o != nil && o.Name != "" {
		records = append(records, record{Type: strings.ToLower(o.Type), Name: o.Name, Value: o.Value})
	}
	errs := d.VerificationErrors
	if errs == nil {
		errs = []string{}
	}

	var diags diag.Diagnostics
	recordList, d1 := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: recordTypes}, records)
	errorList, d2 := types.ListValueFrom(ctx, types.StringType, errs)
	diags.Append(d1...)
	diags.Append(d2...)
	if diags.HasError() {
		return types.ObjectNull(domainStatusTypes), diags
	}
	obj, d3 := types.ObjectValue(domainStatusTypes, map[string]attr.Value{
		"status":               types.StringValue(d.Status),
		"ssl_status":           types.StringValue(d.SSL.Status),
		"verification_records": recordList,
		"verification_errors":  errorList,
	})
	diags.Append(d3...)
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
