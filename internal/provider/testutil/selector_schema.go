package testutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// AssertSingularSelectorSchema verifies the common singular lookup contract.
func AssertSingularSelectorSchema(t *testing.T, dataSource datasource.DataSource, selectorFields ...string) {
	t.Helper()
	if _, ok := dataSource.(datasource.DataSourceWithConfigValidators); !ok {
		t.Fatal("singular data source must validate selectors during configuration")
	}

	var response datasource.SchemaResponse
	dataSource.Schema(context.Background(), datasource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", response.Diagnostics)
	}

	for _, field := range append([]string{"id"}, selectorFields...) {
		attribute, ok := response.Schema.Attributes[field].(dsschema.StringAttribute)
		if !ok {
			t.Fatalf("%s is not a string attribute", field)
		}
		if !attribute.Optional || !attribute.Computed || attribute.Required {
			t.Errorf("%s must be optional and computed, got optional=%t computed=%t required=%t", field, attribute.Optional, attribute.Computed, attribute.Required)
		}
	}
}
