package shared

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	nb "github.com/nautobot/go-nautobot/v3"
)

type APIClient struct {
	Client *nb.APIClient
}

func configuredAPIClient(providerData any) (*APIClient, bool) {
	client, ok := providerData.(*APIClient)
	return client, ok && client != nil
}

func ConfigureAPIClient(providerData any, diagnostics *diag.Diagnostics) *APIClient {
	client, ok := configuredAPIClient(providerData)
	if !ok {
		diagnostics.AddError(
			"Unexpected provider data type",
			"The provider supplied an invalid API client. This is a provider bug.",
		)
		return nil
	}
	return client
}
