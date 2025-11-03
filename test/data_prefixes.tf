data "nautobot_prefixes" "example" {
}

output "prefixes_details" {
  value = data.nautobot_prefixes.example.prefixes[0]
}

output "prefixes_id" {
  value = data.nautobot_prefixes.example.prefixes[0].id
}