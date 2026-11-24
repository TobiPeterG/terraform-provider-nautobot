data "nautobot_namespace" "global" {
  name = "Global"
}

// Fetch a prefix by its exact CIDR and namespace
data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

// Allocate the first available IP address from that prefix
resource "nautobot_available_ip_address" "example" {
  prefix_id   = data.nautobot_prefix.example.id
  status      = "Active"
  type        = "host"
  dns_name    = "test-vm.test.com"
  description = "Allocated for test VM"
  # Optional metadata fields:
  # role_id       = "00000000-0000-0000-0000-000000000001"
  # tenant_id     = "00000000-0000-0000-0000-000000000002"
  # nat_inside_id = "00000000-0000-0000-0000-000000000003"
  # tags_ids      = ["00000000-0000-0000-0000-000000000004"]
}

# An IP may alternatively be allocated from a non-exclusive IP address range.
resource "nautobot_available_ip_address" "from_range" {
  ip_address_range_id = "00000000-0000-0000-0000-000000000000"
  status              = "Active"
}

output "resource_available_ip_address" {
  value = nautobot_available_ip_address.example
}
