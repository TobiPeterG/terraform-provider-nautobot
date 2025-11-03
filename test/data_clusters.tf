data "nautobot_clusters" "example" {
  depends_on = [nautobot_cluster.new]
}

output "clusters_details" {
  value = data.nautobot_clusters.example.clusters[0]
}

output "clusters_id" {
  value = data.nautobot_clusters.example.clusters[0].id
}
