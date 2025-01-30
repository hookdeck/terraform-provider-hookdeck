package sourcev2

import (
	"terraform-provider-hookdeck/internal/provider/sourcev2/generated"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func NewSourceResources() []func() resource.Resource {
	return generated.ResourceFactories()
}
