data "nautobot_vlan_group" "example" {
  name = "Campus VLANs"
}

data "nautobot_vlan" "example" {
  name          = "My VLAN Name"
  vlan_group_id = data.nautobot_vlan_group.example.id
}

output "data_vlan" {
  value = data.nautobot_vlan.example
}
