terraform {
  required_providers {
    nautobot = {
      version = "3.0.5"
      source  = "github.com/TobiPeterG/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://demo.nautobot.com/api"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}
