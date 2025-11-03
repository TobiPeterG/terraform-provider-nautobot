data "nautobot_graphql" "nodes" {
  depends_on = [nautobot_virtual_machine.new]
  query = <<EOF
query {
  virtual_machines {
      name
      id
  }
  devices {
    name
    id
  }
}
EOF
}

output "data_source_graphql" {
  value = data.nautobot_graphql.nodes
}
output "data_source_graphql_vm" {
  value = jsondecode(data.nautobot_graphql.nodes.data).virtual_machines
}
