// This fetches the VLAN with the given name
data "nautobot_vlan" "example" {
  name = "My VLAN Name"
}

// So we can get the first prefix belonging to it
data "nautobot_prefix" "example" {
  vlan_id = data.nautobot_vlan.example.id
}

// And finally get the first available IP address from that prefix
resource "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
  status    = "Active"
  dns_name  = "test-vm.test.com"
}

output "resource_available_ip_address" {
  value = nautobot_available_ip_address.example
}
