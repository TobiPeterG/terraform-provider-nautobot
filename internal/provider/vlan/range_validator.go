package vlan

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type vlanRangeValidator struct{}

func (vlanRangeValidator) Description(context.Context) string {
	return "a comma-separated list of VLAN IDs or ranges between 1 and 4094"
}

func (v vlanRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (vlanRangeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := validateVLANRange(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid VLAN range", err.Error())
	}
}

func validateVLANRange(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("range must not be empty")
	}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		bounds := strings.Split(part, "-")
		if len(bounds) > 2 || len(bounds) == 0 {
			return fmt.Errorf("%q is not a VLAN ID or range", part)
		}
		start, err := parseVLANID(bounds[0])
		if err != nil {
			return err
		}
		if len(bounds) == 2 {
			end, err := parseVLANID(bounds[1])
			if err != nil {
				return err
			}
			if start > end {
				return fmt.Errorf("VLAN range %q starts after it ends", part)
			}
		}
	}
	return nil
}

func parseVLANID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id < 1 || id > 4094 {
		return 0, fmt.Errorf("%q is not a VLAN ID between 1 and 4094", value)
	}
	return id, nil
}
