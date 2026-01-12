terraform {
  required_providers {
    nautobot = {
      version = "3.0.1"
      source  = "registry.terraform.io/nautobot/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://demo.nautobot.com/api"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
