data "nautobot_cluster" "example" {
  depends_on = [nautobot_cluster.new]
  name = nautobot_cluster.new.name
}

output "cluster_details" {
  value = data.nautobot_cluster.example
}

output "cluster_id" {
  value = data.nautobot_cluster.example.id
}