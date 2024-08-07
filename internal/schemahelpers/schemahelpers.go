package schemahelpers

import (
	"log"

	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DataSourceSchemaFromResourceSchema(resourceSchema map[string]resource_schema.Attribute, idField string) map[string]datasource_schema.Attribute {
	foundIdField := false

	datasourceSchema := make(map[string]datasource_schema.Attribute)
	for name, srcAttr := range resourceSchema {
		optional := false
		required := false
		computed := true

		if name == idField {
			required = true
			computed = false
			foundIdField = true
		}

		switch srcAttrTyped := srcAttr.(type) {
		case resource_schema.StringAttribute:
			datasourceSchema[name] = datasource_schema.StringAttribute{
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.BoolAttribute:
			datasourceSchema[name] = datasource_schema.BoolAttribute{
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.Int64Attribute:
			datasourceSchema[name] = datasource_schema.Int64Attribute{
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.ListAttribute:
			datasourceSchema[name] = datasource_schema.ListAttribute{
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				ElementType:         srcAttrTyped.ElementType,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.MapAttribute:
			datasourceSchema[name] = datasource_schema.MapAttribute{
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				ElementType:         srcAttrTyped.ElementType,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.SetAttribute:
			datasourceSchema[name] = datasource_schema.SetAttribute{
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				ElementType:         srcAttrTyped.ElementType,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.ObjectAttribute:
			datasourceSchema[name] = datasource_schema.ObjectAttribute{
				AttributeTypes:      srcAttrTyped.AttributeTypes,
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		case resource_schema.SingleNestedAttribute:
			datasourceSchema[name] = datasource_schema.SingleNestedAttribute{
				Attributes:          DataSourceSchemaFromResourceSchema(srcAttrTyped.Attributes, ""),
				Validators:          srcAttrTyped.Validators,
				Description:         srcAttrTyped.Description,
				MarkdownDescription: srcAttrTyped.MarkdownDescription,
				CustomType:          srcAttrTyped.CustomType,
				Sensitive:           srcAttrTyped.Sensitive,
				Optional:            optional,
				Required:            required,
				Computed:            computed,
			}
		default:
			log.Panicf("unknown attribute type: %v", srcAttr.GetType().String())
		}
	}

	if !foundIdField && idField != "" {
		log.Panicf("id field \"%s\" not found in resource schema", idField)
	}

	return datasourceSchema
}
