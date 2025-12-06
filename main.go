package main

import (
	"context"

	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider"
)

var version = "dev"

func main() {
	providerserver.Serve(
		context.Background(),
		func() tfprovider.Provider { return provider.New(version) },
		providerserver.ServeOpts{
			Address: "registry.terraform.io/nautobot/nautobot",
		},
	)
}
