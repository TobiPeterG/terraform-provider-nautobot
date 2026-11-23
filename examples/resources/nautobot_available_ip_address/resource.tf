data "nautobot_namespace" "global" {
  name = "Global"
}

// Fetch a prefix by its exact CIDR and namespace
data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

// Allocate the next available IP address from that prefix
resource "nautobot_available_ip_address" "example" {
  prefix_id   = data.nautobot_prefix.example.id
  status      = "Active"
  type        = "host"
  dns_name    = "test-vm.test.com"
  description = "Allocated for test VM"
}

output "resource_available_ip_address" {
  value = nautobot_available_ip_address.example
}
