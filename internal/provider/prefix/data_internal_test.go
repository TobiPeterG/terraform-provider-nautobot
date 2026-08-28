package prefix

import (
	"testing"

	nb "github.com/nautobot/go-nautobot/v3"
)

func TestPrefixModelFromAPIRejectsMalformedResponses(t *testing.T) {
	t.Parallel()
	statusName := func(string) (string, error) { return "", nil }
	if _, err := prefixModelFromAPI(nil, statusName); err == nil {
		t.Fatal("expected nil response to fail")
	}
	if _, err := prefixModelFromAPI(&nb.Prefix{}, statusName); err == nil {
		t.Fatal("expected response without ID to fail")
	}
}
