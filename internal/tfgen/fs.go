package tfgen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-codegen-spec/spec"
)

func writeTFCodeSpec(spec spec.Specification) error {
	bytes, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshalling provider code spec to JSON: %w", err)
	}
	return os.WriteFile("../../assets/provider-code-spec.json", bytes, 0644)
}
