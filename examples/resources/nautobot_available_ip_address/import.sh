#!/usr/bin/env sh
# Available IPs can be imported using the IP address UUID, e.g. '748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7'
terraform import nautobot_available_ip_address.example "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"

# Or using '<prefix_id>,<ip_id>'
terraform import nautobot_available_ip_address.example "3f8f4b3a-7b31-4f2d-9b36-2d2c3b4c5d6e,748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
