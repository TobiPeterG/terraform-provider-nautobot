// This fetches the VLAN with the given name
data "nautobot_vlan" "example" {
  name = "My VLAN Name"
}

// So we can get the first prefix belonging to it
data "nautobot_prefix" "example" {
  vlan_id = data.nautobot_vlan.example.id
}

// And finally get the first available IP address from that prefix
data "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
}

output "data_available_ip_address" {
  value = data.nautobot_available_ip_address.example
}
