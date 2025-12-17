terraform {
  required_providers {
    nautobot = {
      version = "3.0.0"
      source  = "registry.terraform.io/TobiPeterG/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://my.instance.com/api"
  token = "MyAPIToken0000abcdefghijklmnopqrstuvwxyz"
}
