package shared_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestNullableReferenceJSON(t *testing.T) {
	empty, err := json.Marshal(shared.NullableReference(""))
	if err != nil {
		t.Fatalf("marshal empty FK: %v", err)
	}
	if string(empty) != "null" {
		t.Fatalf("expected empty FK to marshal as null, got %s", empty)
	}
	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	populated, err := json.Marshal(shared.NullableReference(want))
	if err != nil {
		t.Fatalf("marshal populated FK: %v", err)
	}
	if !strings.Contains(string(populated), want) {
		t.Fatalf("expected populated FK JSON to contain %q, got %s", want, populated)
	}
}
