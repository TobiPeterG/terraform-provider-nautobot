terraform {
  required_providers {
    nautobot = {
      version = "3.0.2"
      source  = "github.com/TobiPeterG/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://demo.nautobot.com/api"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
