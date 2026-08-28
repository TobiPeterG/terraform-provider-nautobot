package shared

import (
	"context"
	"fmt"

	nb "github.com/nautobot/go-nautobot/v3"
)

// ResolveAvailableIPSource maps either supported Terraform allocation source to
// Nautobot's prefix available-IP endpoint and its optional range bounds.
func ResolveAvailableIPSource(ctx context.Context, client *nb.APIClient, prefixID, rangeID string) (string, string, string, error) {
	if (prefixID == "") == (rangeID == "") {
		return "", "", "", fmt.Errorf("exactly one of `prefix_id` or `ip_address_range_id` must be provided")
	}
	if rangeID == "" {
		return prefixID, "", "", nil
	}

	rng, httpResp, err := client.IpamAPI.IpamIpAddressRangesRetrieve(ctx, rangeID).Execute()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to retrieve IP address range: %s", HTTPError(err, httpResp))
	}
	if rng.IsExclusive != nil && *rng.IsExclusive {
		return "", "", "", fmt.Errorf("IP address range %s is exclusive and cannot contain individual IP addresses", rangeID)
	}
	if rng.Parent == nil || rng.Parent.Id == nil || rng.Parent.Id.String == nil || *rng.Parent.Id.String == "" {
		return "", "", "", fmt.Errorf("IP address range %s has no parent prefix", rangeID)
	}

	return *rng.Parent.Id.String, rng.StartAddress, rng.EndAddress, nil
}
