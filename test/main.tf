terraform {
  required_providers {
    nautobot = {
      version = "0.0.1-beta"
      source  = "github.com/nautobot/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://demo.nautobot.com/api"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}

resource "nautobot_manufacturer" "new" {
  description = "Created with Terraform"
  name        = "New Vendor"
}

resource "nautobot_cluster_type" "new" {
  name        = "Example Cluster Type"
  description = "This is a cluster type created via Terraform"
}

resource "nautobot_cluster" "new" {
  name            = "My New Cluster"
  comments        = "This cluster was created using Terraform."
  cluster_type_id = nautobot_cluster_type.new.id 

  # Optionally add cluster group, tenant, location, etc.
#  cluster_group_id   = "your-cluster-group-id"
#  tenant_id          = data.nautobot_tenant.example.id  # Referencing tenant data source
#  location_id        = "your-location-id"
#  tags_id            = ["tag1", "tag2"]

  custom_fields = {
    "custom_field_1" = "value1"
    "custom_field_2" = "value2"
  }
}

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
#  primary_ip4_id     = "192.168.0.100"  # Optional
#  primary_ip6_id     = "2001:db8::100"  # Optional
#  software_version_id = "v1.0"          # Optional

  custom_fields = {
    custom_field_1 = "Custom value 1"
    custom_field_2 = "Custom value 2"
  }

#  tags_ids = ["tag1", "tag2"] # Optional tags
}

resource "nautobot_vm_interface" "new" {
  name = "eth0"
  virtual_machine_id = nautobot_virtual_machine.new.id
  status = "Active"
}