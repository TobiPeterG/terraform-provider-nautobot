package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nautobot": providerserver.NewProtocol6WithError(New("test")),
}

func testAccPreCheck(t *testing.T) {
}

func testAccProviderConfig() string {
	url := "https://demo.nautobot.com/api"
	token := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	return `
provider "nautobot" {
  url   = "` + url + `"
  token = "` + token + `"
}
`
}
