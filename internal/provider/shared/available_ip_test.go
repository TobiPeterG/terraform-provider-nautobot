package shared_test

import (
	"context"
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func TestResolveAvailableIPSourceValidation(t *testing.T) {
	t.Parallel()
	if _, _, _, err := shared.ResolveAvailableIPSource(context.Background(), nil, "", ""); err == nil {
		t.Fatal("expected an error when neither allocation source is set")
	}
	if _, _, _, err := shared.ResolveAvailableIPSource(context.Background(), nil, "prefix", "range"); err == nil {
		t.Fatal("expected an error when both allocation sources are set")
	}
	prefix, start, end, err := shared.ResolveAvailableIPSource(context.Background(), nil, "prefix", "")
	if err != nil || prefix != "prefix" || start != "" || end != "" {
		t.Fatalf("unexpected prefix source result: prefix=%q start=%q end=%q err=%v", prefix, start, end, err)
	}
}
