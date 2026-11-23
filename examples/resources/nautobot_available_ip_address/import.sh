#!/usr/bin/env sh
# Available IPs can be imported using the IP UUID (prefix_id is derived from parent), e.g.:
terraform import nautobot_available_ip_address.example "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"

# Or explicitly specifying `<prefix_id>,<ip_id>`:
terraform import nautobot_available_ip_address.example "11111111-2222-3333-4444-555555555555,748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
