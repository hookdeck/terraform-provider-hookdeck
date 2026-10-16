package destination

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Both names read their own configuration: team_id is on the v2 name only.
func TestValidateConfig_bothNames(t *testing.T) {
	for name, r := range map[string]*destinationResource{
		"hookdeck_gateway_destination": newDestinationResource(false),
		"hookdeck_destination":         newDestinationResource(true),
	} {
		t.Run(name, func(t *testing.T) {
			schema := r.schema()
			typ, ok := schema.Type().TerraformType(t.Context()).(tftypes.Object)
			if !ok {
				t.Fatal("schema type is not an object")
			}
			values := map[string]tftypes.Value{}
			for attr, attrType := range typ.AttributeTypes {
				values[attr] = tftypes.NewValue(attrType, nil)
			}
			values["name"] = tftypes.NewValue(tftypes.String, "api")
			values["type"] = tftypes.NewValue(tftypes.String, "HTTP")

			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(t.Context(), resource.ValidateConfigRequest{
				Config: tfsdk.Config{Schema: schema, Raw: tftypes.NewValue(typ, values)},
			}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
		})
	}
}
