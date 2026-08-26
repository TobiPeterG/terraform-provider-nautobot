package shared

import "testing"

func TestStatusResolverName(t *testing.T) {
	t.Parallel()
	resolver := &StatusResolver{names: map[string]string{"id": "Active"}}
	if name, err := resolver.Name("id"); err != nil || name != "Active" {
		t.Fatalf("unexpected resolution: name=%q err=%v", name, err)
	}
	if _, err := resolver.Name("missing"); err == nil {
		t.Fatal("expected an error for an unknown status ID")
	}
}
