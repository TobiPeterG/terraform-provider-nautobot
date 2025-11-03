# Example virtual machine resource
resource "nautobot_virtual_machine" "new" {
  name            = "Example VM"
  cluster_id      = nautobot_cluster.new.id
  status          = "Active"
  vcpus           = 4
  memory          = 8192 # Memory in MB (8GB)
  disk            = 100  # Disk size in GB
  comments        = "This virtual machine was created using Terraform."
#  tenant_id          = "some-tenant-id" # Optional
#  platform_id        = "Linux"          # Optional
#  role_id            = "Web Server"     # Optional
#  primary_ip4_id     = nautobot_available_ip_address.example.id
#  primary_ip6_id     = "2001:db8::100"  # Optional
#  software_version_id = "v1.0"          # Optional

#  tags_ids = ["tag1", "tag2"] # Optional tags
}
