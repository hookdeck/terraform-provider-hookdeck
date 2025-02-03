package codegenv2

type SourceType struct {
	Ref        string
	ConfigRef  string
	NameSnake  string // e.g., "shopify"
	NameCamel  string // e.g., "shopify"
	NamePascal string // e.g., "Shopify"
	NameConfig string // similar to NamePascal but for ThreeDEye it's 3DEye
	TypeEnum   string // e.g., "SHOPIFY"
	Fields     []Field
}

type Field struct {
	NameSnake  string // e.g., "webhook_secret_key"
	NameCamel  string // e.g., "webhookSecretKey"
	NamePascal string // e.g., "WebhookSecretKey"
	Required   bool
	Nullable   bool
	Type       FieldType
}

type FieldType interface {
	isFieldType()
	getFieldType() string
}

type StringField struct {
	IsEnum         bool
	EnumNameString string
	EnumValues     []string
}

type ArrayField struct {
	ItemType FieldType
}

type ObjectField struct {
	Properties []Field
	OneOf      []ObjectField // For union types
	Required   []string
}

func (StringField) isFieldType() {}
func (ArrayField) isFieldType()  {}
func (ObjectField) isFieldType() {}

func (StringField) getFieldType() string { return "StringField" }
func (ArrayField) getFieldType() string  { return "ArrayField" }
func (ObjectField) getFieldType() string { return "ObjectField" }

// HasEnums returns true if any field (or nested field) has enums
func (s SourceType) HasEnums() bool {
	for _, field := range s.Fields {
		if hasEnums(field.Type) {
			return true
		}
	}
	return false
}

func hasEnums(fieldType FieldType) bool {
	switch f := fieldType.(type) {
	case StringField:
		return f.IsEnum
	case ArrayField:
		return hasEnums(f.ItemType)
	case ObjectField:
		for _, prop := range f.Properties {
			if hasEnums(prop.Type) {
				return true
			}
		}
	}
	return false
}
