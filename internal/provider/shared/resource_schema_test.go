package shared_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestOptionalResourceSchemaHelpersUseKnownDefaultsWithoutPlanModifiers(t *testing.T) {
	t.Parallel()

	stringAttribute := shared.OptionalStringWithDefault("description")
	if !stringAttribute.Optional || !stringAttribute.Computed || len(stringAttribute.PlanModifiers) != 0 {
		t.Fatalf("unexpected optional string shape: %#v", stringAttribute)
	}
	var stringResponse defaults.StringResponse
	stringAttribute.Default.DefaultString(context.Background(), defaults.StringRequest{}, &stringResponse)
	if !stringResponse.PlanValue.Equal(types.StringValue("")) {
		t.Fatalf("string default = %v, want empty string", stringResponse.PlanValue)
	}

	listAttribute := shared.OptionalStringListWithDefault("description")
	if !listAttribute.Optional || !listAttribute.Computed || len(listAttribute.PlanModifiers) != 0 {
		t.Fatalf("unexpected optional list shape: %#v", listAttribute)
	}

	setAttribute := shared.OptionalStringSetWithDefault("description")
	if !setAttribute.Optional || !setAttribute.Computed || len(setAttribute.PlanModifiers) != 0 {
		t.Fatalf("unexpected optional set shape: %#v", setAttribute)
	}
}
