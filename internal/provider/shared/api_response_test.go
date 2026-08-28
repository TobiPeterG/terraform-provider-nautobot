package shared_test

import (
	"strings"
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestValidateAPIObjectID(t *testing.T) {
	t.Parallel()

	matching := "object-id"
	other := "other-id"
	if err := shared.ValidateAPIObjectID("object", matching, &matching); err != nil {
		t.Fatalf("matching response ID: %v", err)
	}
	if err := shared.ValidateAPIObjectID("object", "", &matching); err != nil {
		t.Fatalf("response ID without requested ID: %v", err)
	}
	if err := shared.ValidateAPIObjectID("object", matching, nil); err == nil || !strings.Contains(err.Error(), "no id") {
		t.Fatalf("missing response ID error = %v", err)
	}
	if err := shared.ValidateAPIObjectID("object", matching, &other); err == nil || !strings.Contains(err.Error(), "other-id") {
		t.Fatalf("mismatched response ID error = %v", err)
	}
	if err := shared.ValidateReturnedObjectID("object", matching, other); err == nil || !strings.Contains(err.Error(), "other-id") {
		t.Fatalf("mismatched model ID error = %v", err)
	}
}
