package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	nb "github.com/nautobot/go-nautobot/v3"
)

func TestResourcesExcludeObservationalAttributes(t *testing.T) {
	t.Parallel()

	// These values describe API presentation or related-object aggregates. They
	// do not configure the managed object and may change independently of it.
	excluded := []string{
		"display",
		"url",
		"natural_slug",
		"notes_url",
		"last_updated",
		"device_count",
		"virtual_machine_count",
		"prefix_count",
		"vlan_count",
	}

	provider := &nautobotProvider{}
	for _, constructor := range provider.Resources(context.Background()) {
		managedResource := constructor()
		var metadata resource.MetadataResponse
		managedResource.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "nautobot"}, &metadata)

		var schemaResponse resource.SchemaResponse
		managedResource.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
		if schemaResponse.Diagnostics.HasError() {
			t.Fatalf("%s schema returned diagnostics: %v", metadata.TypeName, schemaResponse.Diagnostics)
		}

		for _, name := range excluded {
			if _, exists := schemaResponse.Schema.Attributes[name]; exists {
				t.Errorf("%s resource must not expose observational attribute %q", metadata.TypeName, name)
			}
		}
	}
}

type blockingRoundTripper struct{}

func (blockingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

type successfulStatusRoundTripper struct {
	sawDeadline bool
}

func (r *successfulStatusRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	_, r.sawDeadline = req.Context().Deadline()
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"nautobot-version":"3.0.2"}`)),
		Request:    req,
	}, nil
}

func TestCheckVersionCompatibilityTimeout(t *testing.T) {
	t.Parallel()

	config := nb.NewConfiguration()
	config.Servers[0].URL = "http://nautobot.invalid/api"
	config.HTTPClient = &http.Client{Transport: blockingRoundTripper{}}
	api := nb.NewAPIClient(config)

	const timeout = 20 * time.Millisecond
	provider := &nautobotProvider{version: "3.0.2"}
	start := time.Now()
	err := provider.checkVersionCompatibility(context.Background(), api, timeout)

	if err == nil {
		t.Fatal("expected status request timeout error")
	}
	if !strings.Contains(err.Error(), "Nautobot /api/status/ request timed out after 20ms") {
		t.Fatalf("unexpected error: %s", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("status request timeout took too long: %s", elapsed)
	}
}

func TestCheckVersionCompatibilityTimeoutDisabled(t *testing.T) {
	t.Parallel()

	transport := &successfulStatusRoundTripper{}
	config := nb.NewConfiguration()
	config.Servers[0].URL = "http://nautobot.invalid/api"
	config.HTTPClient = &http.Client{Transport: transport}
	api := nb.NewAPIClient(config)

	provider := &nautobotProvider{version: "3.0.2"}
	if err := provider.checkVersionCompatibility(context.Background(), api, 0); err != nil {
		t.Fatalf("unexpected version compatibility error: %s", err)
	}
	if transport.sawDeadline {
		t.Fatal("status request had a deadline when timeout was disabled")
	}
}
