package shared

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SelectorConfigValidator struct{ spec SelectorSpec }

func (v SelectorConfigValidator) Description(context.Context) string {
	return "requires either an ID or a complete natural key"
}

func (v SelectorConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v SelectorConfigValidator) ValidateDataSource(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	read := func(name string) (types.String, bool) {
		var value types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(name), &value)...)
		return value, value.IsUnknown()
	}

	id, unknown := read("id")
	if resp.Diagnostics.HasError() || unknown {
		return
	}
	values := make(map[string]string, len(v.spec.NaturalKey)+len(v.spec.Qualifiers))
	for _, name := range append(append([]string{}, v.spec.NaturalKey...), v.spec.Qualifiers...) {
		value, unknown := read(name)
		if resp.Diagnostics.HasError() || unknown {
			return
		}
		values[name] = value.ValueString()
	}
	if err := v.spec.Validate(id.ValueString(), values); err != nil {
		resp.Diagnostics.AddError("Invalid object selector", err.Error())
	}
}
