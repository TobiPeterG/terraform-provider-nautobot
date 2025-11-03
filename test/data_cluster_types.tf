data "nautobot_cluster_types" "example" {
  depends_on = [nautobot_cluster_type.new]
}

output "cluster_types_details" {
  value = data.nautobot_cluster_types.example.cluster_types[0]
}

output "cluster_types_id" {
  value = data.nautobot_cluster_types.example.cluster_types[0].id
}
