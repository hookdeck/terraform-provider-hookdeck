package tfgen

import "os"

func writeTFCodeSpec(content []byte) error {
	return os.WriteFile("assets/provider-code-spec.json", content, 0644)
}
