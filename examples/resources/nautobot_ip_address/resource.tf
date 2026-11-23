data "nautobot_namespace" "global" {
  name = "Global"
}

data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

resource "nautobot_ip_address" "example" {
  address      = "10.151.0.10/16"
  status       = "Active"
  namespace_id = data.nautobot_namespace.global.id
  type         = "host"
  dns_name     = "app01.example.com"
  description  = "Primary IP for app01"
}

output "resource_ip_address" {
  value = nautobot_ip_address.example
}
