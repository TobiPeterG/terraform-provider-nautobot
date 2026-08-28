resource "nautobot_ip_address_range" "dhcp_pool" {
  start_address     = "10.0.0.100"
  end_address       = "10.0.0.200"
  status            = "Active"
  name              = "DHCP pool"
  description       = "Addresses managed by DHCP"
  count_as_utilized = true
  is_exclusive      = false
}
