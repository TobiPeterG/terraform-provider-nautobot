package testutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// AssertStringAttributesHaveNoPlanModifiers protects API-derived values that
// may change as a side effect of updating another attribute. Preserving their
// old state in the plan would cause an inconsistent-result-after-apply error.
func AssertStringAttributesHaveNoPlanModifiers(t *testing.T, managedResource resource.Resource, attributeNames ...string) {
	t.Helper()

	var response resource.SchemaResponse
	managedResource.Schema(context.Background(), resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("unexpected resource schema diagnostics: %v", response.Diagnostics)
	}

	for _, name := range attributeNames {
		attribute, ok := response.Schema.Attributes[name].(rschema.StringAttribute)
		if !ok {
			t.Errorf("resource schema attribute %q is not a string attribute", name)
			continue
		}
		if len(attribute.PlanModifiers) != 0 {
			t.Errorf("resource schema attribute %q must remain unknown during updates, but has %d plan modifier(s)", name, len(attribute.PlanModifiers))
		}
	}
}
