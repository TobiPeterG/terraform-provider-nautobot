# No VM's exist per default
data "nautobot_virtual_machines" "example" {
  depends_on = [nautobot_vm_primary_ip.new]
}

output "vms_details" {
  value = data.nautobot_virtual_machines.example.virtual_machines[0]
}

output "vms_id" {
  value = data.nautobot_virtual_machines.example.virtual_machines[0].id
}
