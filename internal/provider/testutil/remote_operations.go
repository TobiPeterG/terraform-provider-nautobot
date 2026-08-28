package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
