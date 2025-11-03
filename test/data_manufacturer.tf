data "nautobot_manufacturer" "example" {
  depends_on = [nautobot_manufacturer.new]
  name = nautobot_manufacturer.new.name
}

output "manufacturer_details" {
  value = data.nautobot_manufacturer.example
}

output "manufacturer_id" {
  value = data.nautobot_manufacturer.example.id
}
