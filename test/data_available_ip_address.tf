data "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
}

output "available_ip_address" {
  value = data.nautobot_available_ip_address.example.address
}

output "available_ip_version" {
  value = data.nautobot_available_ip_address.example.ip_version
}