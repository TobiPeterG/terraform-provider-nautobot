package shared_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestSliceToSetAndSetDiff(t *testing.T) {
	a := shared.SliceToSet([]string{"a", "b", "b", "", "c"})
	b := shared.SliceToSet([]string{"b", "c"})
	diff := shared.SetDiff(a, b)
	if len(diff) != 1 || diff[0] != "a" {
		t.Fatalf("expected diff [a], got %v", diff)
	}
}
