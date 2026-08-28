package shared_test

import (
	"context"
	"os"
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestAccGetStatusIDAndGetStatusName(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC is not set; skipping acceptance test")
	}
	client := testutil.AccAPIClient()
	id, err := shared.GetStatusID(context.Background(), client, testutil.Status)
	if err != nil || id == "" {
		t.Fatalf("GetStatusID = %q, %v", id, err)
	}
	name, err := shared.GetStatusName(context.Background(), client, id)
	if err != nil || name != testutil.Status {
		t.Fatalf("GetStatusName = %q, %v; want %q", name, err, testutil.Status)
	}
}
