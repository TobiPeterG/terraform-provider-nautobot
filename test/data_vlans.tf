data "nautobot_vlans" "example" {
}

output "vlans_details" {
  value = data.nautobot_vlans.example.vlans[0]
}

output "vlans_id" {
  value = data.nautobot_vlans.example.vlans[0].id
}