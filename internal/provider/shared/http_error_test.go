package shared_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestCollectStrings(t *testing.T) {
	in := map[string]any{"name": []any{"bad"}, "nested": map[string]any{"other": []any{"also bad"}}, "empty": ""}
	out := shared.CollectStrings(in)
	if len(out) != 2 || out[0] == "" || out[1] == "" {
		t.Fatalf("expected two non-empty messages, got %v", out)
	}
}

func TestBestJSONMessage(t *testing.T) {
	if got := shared.BestJSONMessage(map[string]any{"detail": "nope"}); got != "nope" {
		t.Fatalf("detail message = %q", got)
	}
	if got := shared.BestJSONMessage(map[string]any{"non_field_errors": []any{"x", "y"}, "name": []any{"ignored"}}); got != "x | y" {
		t.Fatalf("non-field message = %q", got)
	}
	if got := shared.BestJSONMessage(map[string]any{"name": []any{"bad name"}, "vid": []any{"bad vid"}}); got != "bad name | bad vid" {
		t.Fatalf("deterministic field message = %q", got)
	}
}

func TestHTTPErrorTwoLineFormatAndBodyPreserved(t *testing.T) {
	body := `{"detail":"boom"}`
	req, _ := http.NewRequest(http.MethodPost, testutil.URL+"/api/x", nil)
	resp := &http.Response{Status: "400 Bad Request", Body: io.NopCloser(bytes.NewBufferString(body)), Request: req}
	out := shared.HTTPError(errors.New("ignored"), resp)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 || lines[0] != "400 Bad Request" || !strings.Contains(lines[1], "boom") {
		t.Fatalf("unexpected formatted error: %q", out)
	}
	if !strings.Contains(lines[1], "(POST "+testutil.URL+"/api/x)") {
		t.Fatalf("formatted error lacks request information: %q", lines[1])
	}
	preserved, _ := io.ReadAll(resp.Body)
	if string(preserved) != body {
		t.Fatalf("preserved body = %q, want %q", preserved, body)
	}
}

func TestHTTPErrorAsError(t *testing.T) {
	t.Parallel()
	err := shared.HTTPErrorAsError(errors.New("request failed"), nil)
	if err == nil || err.Error() != "request failed" {
		t.Fatalf("formatted HTTP error = %v", err)
	}
}
