package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = RFC3339InstantType{}
	_ basetypes.StringValuableWithSemanticEquals = RFC3339Instant{}
	_ xattr.ValidateableAttribute                = RFC3339Instant{}
)

// RFC3339InstantType stores an RFC3339 timestamp and compares values by the
// instant they represent. This prevents harmless API normalization of offsets
// or fractional-second formatting from causing Terraform drift.
type RFC3339InstantType struct {
	basetypes.StringType
}

func (RFC3339InstantType) Equal(other attr.Type) bool {
	_, ok := other.(RFC3339InstantType)
	return ok
}

func (RFC3339InstantType) String() string {
	return "shared.RFC3339InstantType"
}

func (t RFC3339InstantType) ValueFromString(_ context.Context, value basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return RFC3339Instant{StringValue: value}, nil
}

func (t RFC3339InstantType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	baseValue, err := t.StringType.ValueFromTerraform(ctx, value)
	if err != nil {
		return nil, err
	}

	stringValue, ok := baseValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected RFC3339 value type %T", baseValue)
	}

	return RFC3339Instant{StringValue: stringValue}, nil
}

func (RFC3339InstantType) ValueType(context.Context) attr.Value {
	return RFC3339Instant{}
}

// RFC3339Instant is a nullable Terraform string containing an RFC3339
// timestamp.
type RFC3339Instant struct {
	basetypes.StringValue
}

func NewRFC3339InstantNull() RFC3339Instant {
	return RFC3339Instant{StringValue: basetypes.NewStringNull()}
}

func NewRFC3339InstantValue(value time.Time) RFC3339Instant {
	return RFC3339Instant{StringValue: basetypes.NewStringValue(value.Format(time.RFC3339Nano))}
}

func (v RFC3339Instant) Equal(other attr.Value) bool {
	otherValue, ok := other.(RFC3339Instant)
	return ok && v.StringValue.Equal(otherValue.StringValue)
}

func (RFC3339Instant) Type(context.Context) attr.Type {
	return RFC3339InstantType{}
}

func (v RFC3339Instant) ValidateAttribute(_ context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if _, err := time.Parse(time.RFC3339, v.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid RFC3339 timestamp",
			fmt.Sprintf("Value %q is not a valid RFC3339 timestamp: %s", v.ValueString(), err),
		)
	}
}

func (v RFC3339Instant) StringSemanticEquals(_ context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
	otherValue, ok := other.(RFC3339Instant)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Unexpected timestamp value type", fmt.Sprintf("Expected RFC3339Instant, got %T", other))
		return false, diags
	}

	if v.IsNull() || v.IsUnknown() || otherValue.IsNull() || otherValue.IsUnknown() {
		return v.Equal(otherValue), nil
	}

	current, currentErr := time.Parse(time.RFC3339, v.ValueString())
	updated, updatedErr := time.Parse(time.RFC3339, otherValue.ValueString())
	if currentErr != nil || updatedErr != nil {
		return false, nil
	}

	return current.Equal(updated), nil
}
