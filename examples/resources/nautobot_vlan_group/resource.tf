resource "nautobot_vlan_group" "example" {
  name        = "Campus VLANs"
  description = "VLANs used on the campus network"
  range       = "1-999,1100-4094"
}
