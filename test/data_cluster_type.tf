data "nautobot_cluster_type" "example" {
  depends_on = [nautobot_cluster_type.new]
  name = nautobot_cluster_type.new.name
}

output "cluster_type_details" {
  value = data.nautobot_cluster_type.example
}

output "cluster_type_id" {
  value = data.nautobot_cluster_type.example.id
}
