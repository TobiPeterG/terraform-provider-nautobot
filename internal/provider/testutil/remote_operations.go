package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	nb "github.com/nautobot/go-nautobot/v3"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
)

func CaptureResourceID(resourceAddress string, target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceAddress]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddress)
		}
		if resourceState.Primary.ID == "" {
			return fmt.Errorf("%s has an empty ID", resourceAddress)
		}
		*target = resourceState.Primary.ID
		return nil
	}
}

func CheckTenantAbsent(id *string) resource.TestCheckFunc {
	return checkResourceAbsent("tenant", "tenancy/tenants", id)
}

func CheckTenantGroupAbsent(id *string) resource.TestCheckFunc {
	return checkResourceAbsent("tenant group", "tenancy/tenant-groups", id)
}

func checkResourceAbsent(objectName, endpoint string, id *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if *id == "" {
			return fmt.Errorf("%s ID was not captured", objectName)
		}
		response, err := accAPIRequest(http.MethodGet, endpoint+"/"+*id+"/", nil)
		if err != nil {
			return fmt.Errorf("retrieve %s %s: %w", objectName, *id, err)
		}
		defer response.Body.Close()
		if response.StatusCode == http.StatusNotFound {
			return nil
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return fmt.Errorf("retrieve %s %s: HTTP %s", objectName, *id, response.Status)
		}
		return fmt.Errorf("%s %s still exists", objectName, *id)
	}
}

func DeleteResourceOutOfBand(resourceAddress, endpoint string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceAddress]
		if !ok || resourceState.Primary.ID == "" {
			return fmt.Errorf("cannot delete missing resource: %s", resourceAddress)
		}
		response, err := accAPIRequest(http.MethodDelete, endpoint+"/"+resourceState.Primary.ID+"/", nil)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(response.Body)
			return fmt.Errorf("delete %s: HTTP %s: %s", resourceState.Primary.ID, response.Status, strings.TrimSpace(string(body)))
		}
		return nil
	}
}

func MutateResourceOutOfBand(t *testing.T, id, endpoint string, values map[string]any) {
	t.Helper()
	if id == "" {
		t.Fatal("cannot mutate resource: ID was not captured")
	}
	body, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("marshal out-of-band mutation: %v", err)
	}
	response, err := accAPIRequest(http.MethodPatch, endpoint+"/"+id+"/", body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(response.Body)
		t.Fatalf("mutate %s: HTTP %s: %s", id, response.Status, strings.TrimSpace(string(responseBody)))
	}
}

func accAPIRequest(method, requestPath string, body []byte) (*http.Response, error) {
	request, err := http.NewRequestWithContext(context.Background(), method, URL+"/api/"+strings.TrimLeft(requestPath, "/"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Nautobot API request: %w", err)
	}
	request.Header.Set("Authorization", "Token "+Token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute Nautobot API request: %w", err)
	}
	return response, nil
}

func CreateTestTag(t *testing.T, name string, contentTypes []string) string {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	payload, err := json.Marshal(map[string]any{
		"name":          name,
		"content_types": contentTypes,
	})
	if err != nil {
		t.Fatalf("marshal tag %s: %v", name, err)
	}

	resp, err := accAPIRequest(http.MethodPost, "extras/tags/", payload)
	if err != nil {
		t.Fatalf("create tag %s: %v", name, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("create tag %s: HTTP %s: %s", name, resp.Status, strings.TrimSpace(string(body)))
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &created); err != nil || created.ID == "" {
		t.Fatalf("create tag %s: failed to decode ID from %q: %v", name, string(body), err)
	}
	id := created.ID

	t.Cleanup(func() {
		r, delErr := accAPIRequest(http.MethodDelete, "extras/tags/"+id+"/", nil)
		if delErr != nil {
			t.Logf("cleanup tag %s (%s): %v", name, id, delErr)
			return
		}
		defer r.Body.Close()
		if r.StatusCode != http.StatusNoContent && r.StatusCode != http.StatusNotFound {
			b, _ := io.ReadAll(r.Body)
			t.Logf("cleanup tag %s (%s): HTTP %s: %s", name, id, r.Status, strings.TrimSpace(string(b)))
		}
	})

	return id
}

func CreateTestRole(t *testing.T, name string, contentTypes []string) string {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	client := AccAPIClient()
	ctx := context.Background()

	role, resp, err := client.ExtrasAPI.
		ExtrasRolesCreate(ctx).
		RoleRequest(nb.RoleRequest{
			Name:         name,
			ContentTypes: contentTypes,
		}).
		Execute()
	if err != nil {
		t.Fatalf("create role %s: %s", name, shared.HTTPError(err, resp))
	}
	if role.Id == nil || *role.Id == "" {
		t.Fatalf("create role %s: returned empty ID", name)
	}
	id := *role.Id

	t.Cleanup(func() {
		r, delErr := client.ExtrasAPI.ExtrasRolesDestroy(context.Background(), id).Execute()
		if delErr != nil && !shared.IsNotFoundResponse(r) {
			t.Logf("cleanup role %s (%s): %s", name, id, shared.HTTPError(delErr, r))
		}
	})

	return id
}

func CreateTestNamespaceAndPrefix(t *testing.T, nsName, cidr string) (string, string) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	client := AccAPIClient()
	ctx := context.Background()

	ns, resp, err := client.IpamAPI.
		IpamNamespacesCreate(ctx).
		NamespaceRequest(nb.NamespaceRequest{Name: nsName}).
		Execute()
	if err != nil {
		t.Fatalf("create namespace %s: %s", nsName, shared.HTTPError(err, resp))
	}
	if ns.Id == nil || *ns.Id == "" {
		t.Fatalf("create namespace %s: returned empty ID", nsName)
	}
	nsID := *ns.Id

	t.Cleanup(func() {
		r, delErr := client.IpamAPI.IpamNamespacesDestroy(context.Background(), nsID).Execute()
		if delErr != nil && !shared.IsNotFoundResponse(r) {
			t.Logf("cleanup namespace %s (%s): %s", nsName, nsID, shared.HTTPError(delErr, r))
		}
	})

	statusID, err := shared.GetStatusID(ctx, client, Status)
	if err != nil {
		t.Fatalf("GetStatusID(%s): %v", Status, err)
	}

	pfx, resp, err := client.IpamAPI.
		IpamPrefixesCreate(ctx).
		WritablePrefixRequest(nb.WritablePrefixRequest{
			Prefix:    cidr,
			Status:    shared.APIReference(statusID),
			Namespace: shared.APIReferencePointer(nsID),
		}).
		Execute()
	if err != nil {
		t.Fatalf("create prefix %s in namespace %s: %s", cidr, nsName, shared.HTTPError(err, resp))
	}
	if pfx.Id == nil || *pfx.Id == "" {
		t.Fatalf("create prefix %s: returned empty ID", cidr)
	}
	pfxID := *pfx.Id

	t.Cleanup(func() {
		r, delErr := client.IpamAPI.IpamPrefixesDestroy(context.Background(), pfxID).Execute()
		if delErr != nil && !shared.IsNotFoundResponse(r) {
			t.Logf("cleanup prefix %s (%s): %s", cidr, pfxID, shared.HTTPError(delErr, r))
		}
	})

	return nsID, pfxID
}
