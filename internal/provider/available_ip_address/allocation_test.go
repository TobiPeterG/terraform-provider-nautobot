package available_ip_address

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	nb "github.com/nautobot/go-nautobot/v3"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestIsDuplicateAllocationError(t *testing.T) {
	t.Parallel()
	duplicate := "IP address with this Parent and Host already exists"
	if !isDuplicateAllocationError(&http.Response{StatusCode: http.StatusBadRequest}, duplicate) {
		t.Fatal("expected duplicate allocation response to be retryable")
	}
	if isDuplicateAllocationError(&http.Response{StatusCode: http.StatusInternalServerError}, duplicate) {
		t.Fatal("server error must not be classified as duplicate allocation")
	}
	if isDuplicateAllocationError(&http.Response{StatusCode: http.StatusBadRequest}, "another validation error") {
		t.Fatal("unrelated validation error must not be classified as duplicate allocation")
	}
	if isDuplicateAllocationError(nil, duplicate) {
		t.Fatal("missing HTTP response must not be classified as duplicate allocation")
	}
}

func TestResolveAllocatedIPAddressByAddressAndParent(t *testing.T) {
	t.Parallel()

	const (
		address  = "21.45.143.49/28"
		parentID = "66f4da22-7aa5-4e3d-8a9a-723289ea9520"
		resultID = "8358a73c-e33d-4fe7-90b6-ae625eacb534"
	)
	transport := roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(request.URL.Path, "/ipam/ip-addresses/") {
			t.Errorf("request path = %q", request.URL.Path)
		}
		if got := request.URL.Query()["address"]; len(got) != 1 || got[0] != address {
			t.Errorf("address filter = %v", got)
		}
		if got := request.URL.Query()["parent"]; len(got) != 1 || got[0] != parentID {
			t.Errorf("parent filter = %v", got)
		}
		body := `{"count":1,"next":null,"previous":null,"results":[{` +
			`"id":"` + resultID + `",` +
			`"object_type":"ipam.ipaddress",` +
			`"display":"` + address + `",` +
			`"url":"/api/ipam/ip-addresses/` + resultID + `/",` +
			`"natural_slug":"` + address + `",` +
			`"address":"` + address + `",` +
			`"host":"21.45.143.49",` +
			`"mask_length":28,` +
			`"ip_version":4,` +
			`"status":{"id":"9d8b6eba-d27d-4b6f-a9b0-8dc8bb3dad6f"},` +
			`"created":null,"last_updated":null,"notes_url":""}]}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})

	configuration := nb.NewConfiguration()
	configuration.Servers = nb.ServerConfigurations{{URL: "http://nautobot.test"}}
	configuration.HTTPClient = &http.Client{Transport: transport}
	managedResource := &AvailableIPAddressResource{client: &shared.APIClient{Client: nb.NewAPIClient(configuration)}}
	resolved, err := managedResource.resolveAllocatedIPAddress(context.Background(), parentID, &nb.IPAddress{Address: address})
	if err != nil {
		t.Fatalf("resolve allocation: %v", err)
	}
	if resolved.Id == nil || *resolved.Id != resultID {
		t.Fatalf("resolved ID = %v, want %q", resolved.Id, resultID)
	}
}

func TestRandomBackoffBounds(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		attempt int
		minimum time.Duration
		maximum time.Duration
	}{
		{attempt: -1, minimum: 250 * time.Millisecond, maximum: 375 * time.Millisecond},
		{attempt: 1, minimum: 500 * time.Millisecond, maximum: 750 * time.Millisecond},
		{attempt: 20, minimum: 20 * time.Second, maximum: 30 * time.Second},
	} {
		got := randomBackoff(test.attempt)
		if got < test.minimum || got >= test.maximum {
			t.Fatalf("randomBackoff(%d) = %s, want [%s, %s)", test.attempt, got, test.minimum, test.maximum)
		}
	}
}
