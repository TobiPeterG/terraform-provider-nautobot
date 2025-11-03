data "nautobot_vlan" "example" {
  name = "VOICE - wna01-asw-04"
}

output "vlan_details" {
  value = data.nautobot_vlan.example
}

output "vlan_id" {
  value = data.nautobot_vlan.example.id
}
