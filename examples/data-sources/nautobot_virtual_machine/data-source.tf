data "nautobot_virtual_machine" "example" {
  name       = "My VM Name"
  cluster_id = "00000000-0000-0000-0000-000000000000"
}

output "data_virtual_machine" {
  value = data.nautobot_virtual_machine.example
}
