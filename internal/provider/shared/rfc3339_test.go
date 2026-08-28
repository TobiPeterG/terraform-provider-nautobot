package shared

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestRFC3339InstantSemanticEquality(t *testing.T) {
	t.Parallel()

	value := func(input string) RFC3339Instant {
		return RFC3339Instant{StringValue: basetypes.NewStringValue(input)}
	}

	tests := []struct {
		name  string
		left  string
		right string
		equal bool
	}{
		{name: "same value", left: "2026-01-02T03:04:05Z", right: "2026-01-02T03:04:05Z", equal: true},
		{name: "equivalent offset", left: "2026-01-02T04:04:05+01:00", right: "2026-01-02T03:04:05Z", equal: true},
		{name: "equivalent fractional precision", left: "2026-01-02T03:04:05.1200Z", right: "2026-01-02T03:04:05.12Z", equal: true},
		{name: "different instant", left: "2026-01-02T03:04:05Z", right: "2026-01-02T03:04:06Z", equal: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			equal, diags := value(test.left).StringSemanticEquals(context.Background(), value(test.right))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if equal != test.equal {
				t.Fatalf("semantic equality = %t, want %t", equal, test.equal)
			}
		})
	}
}

func TestNewRFC3339InstantValuePreservesFractionalSeconds(t *testing.T) {
	t.Parallel()

	parsed, err := time.Parse(time.RFC3339Nano, "2026-01-02T03:04:05.123456Z")
	if err != nil {
		t.Fatal(err)
	}

	if got := NewRFC3339InstantValue(parsed).ValueString(); got != "2026-01-02T03:04:05.123456Z" {
		t.Fatalf("formatted value = %q", got)
	}
}

func TestRFC3339InstantValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     RFC3339Instant
		wantError bool
	}{
		{name: "valid", value: RFC3339Instant{StringValue: basetypes.NewStringValue("2026-01-02T03:04:05.123Z")}},
		{name: "empty", value: RFC3339Instant{StringValue: basetypes.NewStringValue("")}, wantError: true},
		{name: "malformed", value: RFC3339Instant{StringValue: basetypes.NewStringValue("2026-01-02")}, wantError: true},
		{name: "null", value: NewRFC3339InstantNull()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response xattr.ValidateAttributeResponse
			test.value.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{Path: path.Root("date_allocated")},
				&response,
			)
			if response.Diagnostics.HasError() != test.wantError {
				t.Fatalf("HasError() = %t, want %t: %v", response.Diagnostics.HasError(), test.wantError, response.Diagnostics)
			}
		})
	}
}
