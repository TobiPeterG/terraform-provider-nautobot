package shared

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OptionalStringWithDefault describes an API field where omission and the
// empty string have the same meaning. The static default keeps the planned and
// refreshed representations identical without a plan modifier.
func OptionalStringWithDefault(description string) rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString(""),
		Description: description,
	}
}

// OptionalStringListWithDefault normalizes an omitted list to an empty list so
// API responses cannot cause null-versus-empty drift.
func OptionalStringListWithDefault(description string) rschema.ListAttribute {
	return rschema.ListAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
		Description: description,
	}
}

// OptionalStringSetWithDefault is the set equivalent of
// OptionalStringListWithDefault.
func OptionalStringSetWithDefault(description string) rschema.SetAttribute {
	return rschema.SetAttribute{
		Optional:    true,
		Computed:    true,
		ElementType: types.StringType,
		Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
		Description: description,
	}
}
