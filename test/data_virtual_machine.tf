# No VM's exist per default
data "nautobot_virtual_machine" "example" {
  depends_on = [nautobot_vm_primary_ip.new]
  name = nautobot_virtual_machine.new.name
}

output "vm_details" {
  value = data.nautobot_virtual_machine.example
}

output "vm_id" {
  value = data.nautobot_virtual_machine.example.id
}