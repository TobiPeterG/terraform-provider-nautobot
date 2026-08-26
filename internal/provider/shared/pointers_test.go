package shared_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestStringPtr(t *testing.T) {
	p := shared.StringPtr("x")
	if p == nil || *p != "x" {
		t.Fatalf("expected ptr to %q, got %+v", "x", p)
	}
}

func TestInt32Ptr(t *testing.T) {
	p := shared.Int32Ptr(42)
	if p == nil || *p != int32(42) {
		t.Fatalf("expected ptr to %d, got %+v", 42, p)
	}
}

func TestDerefStr(t *testing.T) {
	if got := shared.DerefString(nil); got != "" {
		t.Fatalf("expected nil pointer to become an empty string, got %q", got)
	}
	empty := ""
	if got := shared.DerefString(&empty); got != "" {
		t.Fatalf("expected empty string to remain empty, got %q", got)
	}
	want := "value"
	if got := shared.DerefString(&want); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
