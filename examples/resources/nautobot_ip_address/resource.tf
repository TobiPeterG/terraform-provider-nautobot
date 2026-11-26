data "nautobot_namespace" "global" {
  name = "Global"
}

data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/24"
  namespace_id = data.nautobot_namespace.global.id
}

resource "nautobot_ip_address" "example" {
  address      = "10.151.0.10/24"
  status       = "Active"
  parent_id    = data.nautobot_prefix.example.id
  namespace_id = data.nautobot_namespace.global.id
  type         = "host"
  dns_name     = "app01.example.com"
  description  = "Static IP address for app01"
  # Optional metadata fields:
  # role_id       = "00000000-0000-0000-0000-000000000001"
  # tenant_id     = "00000000-0000-0000-0000-000000000002"
  # nat_inside_id = "00000000-0000-0000-0000-000000000003"
  # tags_ids      = ["00000000-0000-0000-0000-000000000004"]
}

output "resource_ip_address" {
  value = nautobot_ip_address.example
}
