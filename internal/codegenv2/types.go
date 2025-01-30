package codegenv2

type SourceType struct {
	Ref        string
	ConfigRef  string
	NameSnake  string // e.g., "shopify"
	NameCamel  string // e.g., "shopify"
	NamePascal string // e.g., "Shopify"
	NameConfig string // similar to NamePascal but for ThreeDEye it's 3DEye
	TypeEnum   string // e.g., "SHOPIFY"
	Auth       []AuthField
}

type AuthField struct {
	NameSnake      string // e.g., "webhook_secret_key"
	NameCamel      string // e.g., "webhookSecretKey"
	NamePascal     string // e.g., "WebhookSecretKey"
	Required       bool
	Nullable       bool // Whether the field can be null/omitempty
	IsEnum         bool
	EnumNameString string
	EnumValues     []string // Available enum values
}
