// This fetches the VLAN with the given name
data "nautobot_vlan" "example" {
  name = "My VLAN Name"
}

// So we can get the first prefix belonging to it
data "nautobot_prefix" "example" {
  vlan_id = data.nautobot_vlan.example.id
}

output "data_prefix" {
  value = data.nautobot_prefix.example
}

// We can also get the parent Prefix
data "nautobot_prefix" "example_parent" {
  id = data.nautobot_prefix.example.parent_id
}

output "data_prefix_parent" {
  value = data.nautobot_prefix.example_parent
}