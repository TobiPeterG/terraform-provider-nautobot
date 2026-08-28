package provider

import (
	"net/http"
	"testing"
)

const (
	testURL   = "http://nautobot.example"
	testToken = "test-token"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAuthRoundTripperDoesNotMutateRequest(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, testURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("X-Test", "original")

	transport := &authRoundTripper{
		base: roundTripFunc(func(got *http.Request) (*http.Response, error) {
			if authorization := got.Header.Get("Authorization"); authorization != "Token "+testToken {
				t.Fatalf("unexpected Authorization header %q", authorization)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
		token: testToken,
	}
	if _, err := transport.RoundTrip(request); err != nil {
		t.Fatal(err)
	}
	if authorization := request.Header.Get("Authorization"); authorization != "" {
		t.Fatalf("input request was mutated: Authorization=%q", authorization)
	}
}
