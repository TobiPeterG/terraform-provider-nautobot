package shared_test

import (
	"strings"
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestValidateIDOrNaturalKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		id         string
		naturalKey []shared.SelectorField
		qualifiers []shared.SelectorField
		wantError  string
	}{
		{name: "id", id: "object-id", naturalKey: []shared.SelectorField{{Name: "name"}}},
		{name: "natural key", naturalKey: []shared.SelectorField{{Name: "prefix", Value: "10.0.0.0/24"}, {Name: "namespace_id", Value: "namespace-id"}}},
		{name: "optional qualifier absent", naturalKey: []shared.SelectorField{{Name: "name", Value: "vm"}, {Name: "cluster_id", Value: "cluster-id"}}, qualifiers: []shared.SelectorField{{Name: "tenant_id"}}},
		{name: "optional qualifier present", naturalKey: []shared.SelectorField{{Name: "name", Value: "vm"}, {Name: "cluster_id", Value: "cluster-id"}}, qualifiers: []shared.SelectorField{{Name: "tenant_id", Value: "tenant-id"}}},
		{name: "missing selector", naturalKey: []shared.SelectorField{{Name: "name"}}, wantError: "provide either `id`, or `name`"},
		{name: "incomplete natural key", naturalKey: []shared.SelectorField{{Name: "name", Value: "vm"}, {Name: "cluster_id"}}, wantError: "missing `cluster_id`"},
		{name: "conflicting selector", id: "object-id", naturalKey: []shared.SelectorField{{Name: "name", Value: "vm"}}, wantError: "`id` cannot be combined"},
		{name: "conflicting qualifier", id: "object-id", naturalKey: []shared.SelectorField{{Name: "name"}}, qualifiers: []shared.SelectorField{{Name: "tenant_id", Value: "tenant-id"}}, wantError: "`id` cannot be combined"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := shared.ValidateIDOrNaturalKey(test.id, test.naturalKey, test.qualifiers...)
			if test.wantError == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)) {
				t.Fatalf("expected error containing %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestExactMatchError(t *testing.T) {
	t.Parallel()
	if err := shared.ExactMatchError("prefix", "a natural key", 1); err != nil {
		t.Fatalf("unexpected one-match error: %v", err)
	}
	if err := shared.ExactMatchError("prefix", "a natural key", 0); err == nil || !strings.Contains(err.Error(), "no prefix found") {
		t.Fatalf("unexpected zero-match error: %v", err)
	}
	if err := shared.ExactMatchError("prefix", "a natural key", 2); err == nil || !strings.Contains(err.Error(), "found 2") {
		t.Fatalf("unexpected multiple-match error: %v", err)
	}
}

func TestSelectorSpecValidate(t *testing.T) {
	t.Parallel()
	spec := shared.SelectorSpec{
		NaturalKey: []string{"name", "container_id"},
		Qualifiers: []string{"tenant_id"},
	}
	if err := spec.Validate("", map[string]string{"name": "item", "container_id": "container"}); err != nil {
		t.Fatalf("complete natural key: %v", err)
	}
	if err := spec.Validate("id", map[string]string{"tenant_id": "tenant"}); err == nil || !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("ID with qualifier error = %v", err)
	}
	if got := len(spec.ConfigValidators(nil)); got != 1 {
		t.Fatalf("ConfigValidators count = %d, want 1", got)
	}
}
