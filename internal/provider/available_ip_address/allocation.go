package available_ip_address

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	nb "github.com/nautobot/go-nautobot/v3"
)

// Remove the retry workaround when https://github.com/nautobot/nautobot/issues/8297 is fixed.
const allocationMaxRetries = 10

func randomBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	const (
		base     = 250 * time.Millisecond
		capDelay = 20 * time.Second
	)
	backoff := base << attempt
	if backoff > capDelay {
		backoff = capDelay
	}
	jitterMax := backoff / 2
	if jitterMax <= 0 {
		return backoff
	}
	return backoff + time.Duration(rand.Int63n(int64(jitterMax)))
}

func isDuplicateAllocationError(response *http.Response, formattedError string) bool {
	return response != nil &&
		response.StatusCode == http.StatusBadRequest &&
		strings.Contains(formattedError, "IP address with this Parent and Host already exists")
}

func (r *AvailableIPAddressResource) allocateIPAddress(
	ctx context.Context,
	prefixID, rangeStart, rangeEnd, statusID, dnsName string,
) (*nb.IPAddress, error) {
	body := []nb.IPAllocationRequest{{Status: shared.APIReference(statusID)}}
	if dnsName != "" {
		body[0].DnsName = &dnsName
	}

	for attempt := 0; attempt < allocationMaxRetries; attempt++ {
		request := r.client.Client.IpamAPI.
			IpamPrefixesAvailableIpsCreate(ctx, prefixID).
			IPAllocationRequest(body)
		if rangeStart != "" {
			request = request.RangeStart(rangeStart).RangeEnd(rangeEnd)
		}
		allocated, httpResponse, err := request.Execute()
		if err == nil {
			if len(allocated) == 0 {
				return nil, fmt.Errorf("Nautobot returned no available IP addresses for the selected allocation source")
			}
			if allocated[0].Id != nil && *allocated[0].Id != "" {
				return &allocated[0], nil
			}
			return r.resolveAllocatedIPAddress(ctx, prefixID, &allocated[0])
		}

		formattedError := shared.HTTPError(err, httpResponse)
		if !isDuplicateAllocationError(httpResponse, formattedError) {
			return nil, fmt.Errorf("%s", formattedError)
		}
		if attempt == allocationMaxRetries-1 {
			return nil, fmt.Errorf("duplicate allocation persisted after %d attempts: %s", allocationMaxRetries, formattedError)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while retrying after duplicate allocation: %w", ctx.Err())
		case <-time.After(randomBackoff(attempt)):
		}
	}

	return nil, fmt.Errorf("allocation retry loop ended unexpectedly")
}

func (r *AvailableIPAddressResource) resolveAllocatedIPAddress(
	ctx context.Context,
	prefixID string,
	allocation *nb.IPAddress,
) (*nb.IPAddress, error) {
	if allocation == nil || allocation.Address == "" {
		return nil, fmt.Errorf("Nautobot returned an allocated IP address with neither an id nor an address")
	}

	result, httpResponse, err := r.client.Client.IpamAPI.
		IpamIpAddressesList(ctx).
		Address([]string{allocation.Address}).
		Parent([]string{prefixID}).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("resolve allocated IP address %q: %s", allocation.Address, shared.HTTPError(err, httpResponse))
	}
	selector := fmt.Sprintf("address %q in parent prefix %q", allocation.Address, prefixID)
	if err := shared.ExactMatchError("allocated IP address", selector, len(result.Results)); err != nil {
		return nil, err
	}
	if result.Results[0].Id == nil || *result.Results[0].Id == "" {
		return nil, fmt.Errorf("allocated IP address matching %s returned no id", selector)
	}
	return &result.Results[0], nil
}
