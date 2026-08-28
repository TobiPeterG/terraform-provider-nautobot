package testutil

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	nb "github.com/nautobot/go-nautobot/v3"

	provider "github.com/nautobot/terraform-provider-nautobot/internal/provider"
)

var AccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nautobot": providerserver.NewProtocol6WithError(provider.New("test")),
}

func AccPreCheck(t *testing.T) {
	t.Helper()
	parsedURL, err := url.Parse(URL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		t.Fatalf("NAUTOBOT_TEST_URL must be an absolute HTTP(S) URL, got %q", URL)
	}
	if Token == "" {
		t.Fatal("NAUTOBOT_TEST_TOKEN must not be empty")
	}
}

func AccProviderConfig() string {
	return `
provider "nautobot" {
  url   = "` + URL + `/api"
  token = "` + Token + `"
}
`
}

type authRoundTripper struct {
	base  http.RoundTripper
	token string
}

func (a *authRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	cloned := request.Clone(request.Context())
	cloned.Header = request.Header.Clone()
	cloned.Header.Set("Authorization", "Token "+a.token)
	return a.base.RoundTrip(cloned)
}

func AccAPIClient() *nb.APIClient {
	cfg := nb.NewConfiguration()
	cfg.Servers = nb.ServerConfigurations{{URL: URL + "/api"}}
	cfg.HTTPClient = &http.Client{Transport: &authRoundTripper{base: http.DefaultTransport, token: Token}}
	return nb.NewAPIClient(cfg)
}

var (
	URL   = strings.TrimRight(environmentValue("NAUTOBOT_TEST_URL", "http://nautobot:8080"), "/")
	Token = environmentValue("NAUTOBOT_TEST_TOKEN", "0123456789abcdef0123456789abcdef01234567")
)

const Status = "Active"

func environmentValue(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
