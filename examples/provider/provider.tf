terraform {
  required_providers {
    nautobot = {
      version = "3.0.0-beta"
      source  = "github.com/nautobot/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://my.instance.com/api"
  token = "MyAPIToken0000abcdefghijklmnopqrstuvwxyz"
}
