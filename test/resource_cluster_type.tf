resource "nautobot_cluster_type" "new" {
  name        = "Example Cluster Type"
  description = "This is a cluster type created via Terraform"
}

resource "nautobot_cluster_type" "new2" {
  name        = "Example Cluster Type 2"
  description = "This is a cluster type created via Terraform"
}