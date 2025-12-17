package provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	nb "github.com/nautobot/go-nautobot/v3"
)

func TestNewSecurityProviderNautobotToken(t *testing.T) {
	sp, err := NewSecurityProviderNautobotToken("abc123")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if sp == nil {
		t.Fatalf("expected non-nil provider")
	}
}

func TestSecurityProviderNautobotToken_Intercept(t *testing.T) {
	sp, _ := NewSecurityProviderNautobotToken("abc123")
	req, _ := http.NewRequest(http.MethodGet, testURL, nil)

	if err := sp.Intercept(context.Background(), req); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	got := req.Header.Get("Authorization")
	if got != "Token abc123" {
		t.Fatalf("expected Authorization header %q, got %q", "Token abc123", got)
	}
}

func TestStringPtr(t *testing.T) {
	p := stringPtr("x")
	if p == nil || *p != "x" {
		t.Fatalf("expected ptr to %q, got %+v", "x", p)
	}
}

func TestInt32Ptr(t *testing.T) {
	p := int32Ptr(42)
	if p == nil || *p != int32(42) {
		t.Fatalf("expected ptr to %d, got %+v", 42, p)
	}
}

func TestSliceToSetAndSetDiff(t *testing.T) {
	a := sliceToSet([]string{"a", "b", "b", "", "c"})
	b := sliceToSet([]string{"b", "c"})

	diff := setDiff(a, b)

	if len(diff) != 1 || diff[0] != "a" {
		t.Fatalf("expected diff [a], got %v", diff)
	}
}

func TestCollectStrings(t *testing.T) {
	in := map[string]any{
		"name": []any{"bad"},
		"nested": map[string]any{
			"other": []any{"also bad"},
		},
		"empty": "",
	}
	out := collectStrings(in)

	if len(out) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(out), out)
	}
	if out[0] == "" || out[1] == "" {
		t.Fatalf("expected non-empty messages, got %v", out)
	}
}

func TestBestJSONMessage(t *testing.T) {
	m1 := map[string]any{"detail": "nope"}
	if got := bestJSONMessage(m1); got != "nope" {
		t.Fatalf("expected %q, got %q", "nope", got)
	}

	m2 := map[string]any{"non_field_errors": []any{"x", "y"}, "name": []any{"ignored"}}
	if got := bestJSONMessage(m2); got != "x | y" {
		t.Fatalf("expected %q, got %q", "x | y", got)
	}

	m3 := map[string]any{"name": []any{"bad name"}, "vid": []any{"bad vid"}}
	got3 := bestJSONMessage(m3)
	if got3 == "" {
		t.Fatalf("expected non-empty message, got empty")
	}
}

func TestHttpErr_TwoLineFormatAndBodyPreserved(t *testing.T) {
	body := `{"detail":"boom"}`
	req, _ := http.NewRequest(http.MethodPost, testURL+"/api/x", nil)
	resp := &http.Response{
		Status:  "400 Bad Request",
		Body:    io.NopCloser(bytes.NewBufferString(body)),
		Request: req,
	}

	out := httpErr(errors.New("ignored"), resp)

	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if lines[0] != "400 Bad Request" {
		t.Fatalf("expected status line %q, got %q", "400 Bad Request", lines[0])
	}
	if !strings.Contains(lines[1], "boom") {
		t.Fatalf("expected message line to contain %q, got %q", "boom", lines[1])
	}
	if !strings.Contains(lines[1], "(POST "+testURL+"/api/x)") {
		t.Fatalf("expected message line to contain request info, got %q", lines[1])
	}

	b2, _ := io.ReadAll(resp.Body)
	if string(b2) != body {
		t.Fatalf("expected body preserved %q, got %q", body, string(b2))
	}
}

func TestAccGetStatusIDAndGetStatusName(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC is not set; skipping acceptance test")
	}

	token := testToken
	baseURL := testURL + "/api"

	cfg := nb.NewConfiguration()
	cfg.Servers = nb.ServerConfigurations{
		{URL: baseURL},
	}

	httpClient := &http.Client{
		Transport: &authRT{
			base:  http.DefaultTransport,
			token: token,
		},
	}
	cfg.HTTPClient = httpClient

	client := nb.NewAPIClient(cfg)
	ctx := context.Background()

	id, err := getStatusID(ctx, client, testStatus)
	if err != nil {
		t.Fatalf("getStatusID: expected nil err, got %v", err)
	}
	if id == "" {
		t.Fatalf("getStatusID: expected non-empty id")
	}

	name, err := getStatusName(ctx, client, id)
	if err != nil {
		t.Fatalf("getStatusName: expected nil err, got %v", err)
	}
	if name != testStatus {
		t.Fatalf("getStatusName: expected %q, got %q", testStatus, name)
	}
}
