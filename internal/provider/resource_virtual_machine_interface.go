package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceVMInterface() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a VM Interface in Nautobot",

		CreateContext: resourceVMInterfaceCreate,
		ReadContext:   resourceVMInterfaceRead,
		UpdateContext: resourceVMInterfaceUpdate,
		DeleteContext: resourceVMInterfaceDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the VM interface.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"mac_address": {
				Description: "MAC address of the interface.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"enabled": {
				Description: "Whether the interface is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"mtu": {
				Description: "MTU size of the interface.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"mode": {
				Description: "Mode of the interface.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"description": {
				Description: "Description of the interface.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"status": {
				Description: "Status of the VM interface.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"virtual_machine_id": {
				Description: "ID of the virtual machine to which the interface belongs.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"untagged_vlan_id": {
				Description: "Untagged VLAN ID associated with the interface.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags_ids": {
				Description: "Tags associated with the interface.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_addresses": {
				Description: "List of IP addresses to assign to the VM interface.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "Creation date of the interface.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Last updated date of the interface.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceVMInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {Key: t, Prefix: "Token"},
		},
	)

	// Check if the interface with the same name and virtual machine ID exists
	interfaceName := d.Get("name").(string)
	virtualMachineID := d.Get("virtual_machine_id").(string)

	existingInterfaces, _, err := c.VirtualizationAPI.VirtualizationInterfacesList(auth).
		Name([]string{interfaceName}).
		VirtualMachine([]string{virtualMachineID}).
		Execute()
	if err != nil {
		return diag.Errorf("failed to list VM interfaces: %s", err.Error())
	}

	if len(existingInterfaces.Results) > 0 {
		// Interface already exists, use its ID and skip creation
		if existingInterfaces.Results[0].Id == nil || *existingInterfaces.Results[0].Id == "" {
			return diag.Errorf("existing VM interface %q returned no id", interfaceName)
		}
		d.SetId(*existingInterfaces.Results[0].Id)
		return resourceVMInterfaceRead(ctx, d, meta)
	}

	// Convert status name to ID
	statusName := d.Get("status").(string)
	statusID, err := getStatusID(ctx, c, t, statusName)
	if err != nil {
		return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
	}

	// Prepare the VMInterface request
	var vmInterface nb.WritableVMInterfaceRequest
	vmInterface.Name = interfaceName
	vmInterface.Status = nb.BulkWritableCableRequestStatus{
		Id: &nb.BulkWritableCableRequestStatusId{String: &statusID},
	}

	// Set optional fields
	if v, ok := d.GetOk("mac_address"); ok && v.(string) != "" {
		vmInterface.MacAddress.Set(stringPtr(v.(string)))
	}
	if v, ok := d.GetOk("enabled"); ok {
		enabled := v.(bool)
		vmInterface.Enabled = &enabled
	}
	if v, ok := d.GetOk("mtu"); ok {
		mtu := int32(v.(int))
		vmInterface.Mtu.Set(&mtu)
	}
	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		desc := v.(string)
		vmInterface.Description = &desc
	}

	// Handle virtual machine ID
	vmInterface.VirtualMachine.Id = &nb.BulkWritableCableRequestStatusId{String: &virtualMachineID}

	// Create the interface
	rsp, _, err := c.VirtualizationAPI.VirtualizationInterfacesCreate(auth).WritableVMInterfaceRequest(vmInterface).Execute()
	if err != nil {
		return diag.Errorf("failed to create VM interface: %s", err.Error())
	}

	// Set resource ID
	if rsp.Id == nil || *rsp.Id == "" {
		return diag.Errorf("created VM interface returned no id")
	}
	d.SetId(*rsp.Id)

	// Assign IP addresses to the VM interface
	if v, ok := d.GetOk("ip_addresses"); ok {
		ipAddresses := v.([]interface{})
		for _, ip := range ipAddresses {
			str := ip.(string)
			if str == "" {
				continue
			}
			if err := assignIPAddressToVMInterface(ctx, c, t, str, *rsp.Id); err != nil {
				return diag.Errorf("failed to assign IP address to VM interface: %s", err.Error())
			}
		}
	}

	return resourceVMInterfaceRead(ctx, d, meta)
}

func resourceVMInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {Key: t, Prefix: "Token"},
		},
	)

	// Fetch interface by ID
	vmInterfaceId := d.Id()
	vmInterface, _, err := c.VirtualizationAPI.VirtualizationInterfacesRetrieve(auth, vmInterfaceId).Execute()
	if err != nil {
		return diag.Errorf("failed to read VM interface: %s", err.Error())
	}

	// Map the retrieved data back to Terraform state
	d.Set("name", vmInterface.Name)

	// mac_address
	if vmInterface.MacAddress.IsSet() && vmInterface.MacAddress.Get() != nil {
		d.Set("mac_address", *vmInterface.MacAddress.Get())
	} else {
		d.Set("mac_address", "")
	}

	// enabled
	if vmInterface.Enabled != nil {
		d.Set("enabled", *vmInterface.Enabled)
	} else {
		d.Set("enabled", false)
	}

	// mtu
	if vmInterface.Mtu.IsSet() && vmInterface.Mtu.Get() != nil {
		d.Set("mtu", int(*vmInterface.Mtu.Get()))
	} else {
		d.Set("mtu", 0)
	}

	// description
	if vmInterface.Description != nil {
		d.Set("description", *vmInterface.Description)
	} else {
		d.Set("description", "")
	}

	// status -> name (default to "")
	statusName := ""
	if vmInterface.Status.Id != nil && vmInterface.Status.Id.String != nil {
		statusID := *vmInterface.Status.Id.String
		if statusID != "" {
			if n, err := getStatusName(ctx, c, t, statusID); err == nil {
				statusName = n
			}
		}
	}
	d.Set("status", statusName)

	// virtual_machine_id -> default ""
	vmID := ""
	if vmInterface.VirtualMachine.Id != nil && vmInterface.VirtualMachine.Id.String != nil {
		vmID = *vmInterface.VirtualMachine.Id.String
	}
	d.Set("virtual_machine_id", vmID)

	// untagged_vlan_id -> default ""
	untagged := ""
	if vmInterface.UntaggedVlan.IsSet() {
		if uv := vmInterface.UntaggedVlan.Get(); uv != nil && uv.Id != nil && uv.Id.String != nil {
			untagged = *uv.Id.String
		}
	}
	d.Set("untagged_vlan_id", untagged)

	// created / last_updated
	createdStr := ""
	if vmInterface.Created.IsSet() && vmInterface.Created.Get() != nil {
		createdStr = vmInterface.Created.Get().Format(time.RFC3339)
	}
	d.Set("created", createdStr)

	lastUpdatedStr := ""
	if vmInterface.LastUpdated.IsSet() && vmInterface.LastUpdated.Get() != nil {
		lastUpdatedStr = vmInterface.LastUpdated.Get().Format(time.RFC3339)
	}
	d.Set("last_updated", lastUpdatedStr)

	// tags_ids -> always set list
	tags := make([]string, 0, len(vmInterface.Tags))
	for _, tag := range vmInterface.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tags = append(tags, *tag.Id.String)
		}
	}
	d.Set("tags_ids", tags)

	// ip_addresses -> always set list
	assignedIPs := []string{}
	for _, ip := range vmInterface.IpAddresses {
		if ip.Id != nil && ip.Id.String != nil {
			assignedIPs = append(assignedIPs, *ip.Id.String)
		}
	}
	d.Set("ip_addresses", assignedIPs)

	// mode -> default ""
	mode := ""
	if vmInterface.Mode != nil {
		if vmInterface.Mode.Label != nil {
			mode = string(*vmInterface.Mode.Label)
		} else if vmInterface.Mode.Value != nil {
			mode = string(*vmInterface.Mode.Value)
		}
	}
	d.Set("mode", mode)

	return nil
}

func resourceVMInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	vmInterfaceId := d.Id()

	var vmInterface nb.PatchedWritableVMInterfaceRequest

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {Key: t, Prefix: "Token"},
		},
	)

	// Update the fields that have changed
	if d.HasChange("name") {
		name := d.Get("name").(string)
		vmInterface.Name = &name
	}
	if d.HasChange("mac_address") {
		mac := d.Get("mac_address").(string)
		if mac == "" {
			vmInterface.MacAddress.Unset()
		} else {
			vmInterface.MacAddress.Set(&mac)
		}
	}
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		vmInterface.Enabled = &enabled
	}
	if d.HasChange("mtu") {
		mtu := int32(d.Get("mtu").(int))
		vmInterface.Mtu.Set(&mtu)
	}
	if d.HasChange("description") {
		description := d.Get("description").(string)
		if description == "" {
			vmInterface.Description = nil
		} else {
			vmInterface.Description = &description
		}
	}
	if d.HasChange("status") {
		statusName := d.Get("status").(string)
		statusID, err := getStatusID(ctx, c, t, statusName)
		if err != nil {
			return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
		}
		vmInterface.Status = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{String: &statusID},
		}
	}
	if d.HasChange("virtual_machine_id") {
		vmID := d.Get("virtual_machine_id").(string)
		vmInterface.VirtualMachine = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{String: &vmID},
		}
	}

	// Call the API to update the VM interface
	_, _, err := c.VirtualizationAPI.VirtualizationInterfacesPartialUpdate(auth, vmInterfaceId).PatchedWritableVMInterfaceRequest(vmInterface).Execute()
	if err != nil {
		return diag.Errorf("failed to update VM interface: %s", err.Error())
	}

	// Update IP addresses if they have changed
	if d.HasChange("ip_addresses") {
		oldIPsRaw, newIPsRaw := d.GetChange("ip_addresses")

		// Remove old IP addresses
		for _, oldIP := range oldIPsRaw.([]interface{}) {
			str := oldIP.(string)
			if str == "" {
				continue
			}
			if err := removeIPAddressFromVMInterface(ctx, c, t, str, vmInterfaceId); err != nil {
				return diag.Errorf("failed to remove IP address from VM interface: %s", err.Error())
			}
		}

		// Assign new IP addresses
		for _, newIP := range newIPsRaw.([]interface{}) {
			str := newIP.(string)
			if str == "" {
				continue
			}
			if err := assignIPAddressToVMInterface(ctx, c, t, str, vmInterfaceId); err != nil {
				return diag.Errorf("failed to assign IP address to VM interface: %s", err.Error())
			}
		}
	}

	return resourceVMInterfaceRead(ctx, d, meta)
}

func resourceVMInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {Key: t, Prefix: "Token"},
		},
	)

	// Delete the interface by ID
	vmInterfaceId := d.Id()
	_, err := c.VirtualizationAPI.VirtualizationInterfacesDestroy(auth, vmInterfaceId).Execute()
	if err != nil {
		return diag.Errorf("failed to delete VM interface: %s", err.Error())
	}

	// Clear the ID
	d.SetId("")

	return nil
}

// Helper function to assign an IP address to a VM interface
func assignIPAddressToVMInterface(ctx context.Context, c *nb.APIClient, token, ipAddressID, vmInterfaceID string) error {
	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {Key: token, Prefix: "Token"},
		},
	)

	ipAddressStatusId := nb.BulkWritableCableRequestStatusId{String: &ipAddressID}
	vmInterfaceStatusId := nb.BulkWritableCableRequestStatusId{String: &vmInterfaceID}

	vmInterfaceTenant := nb.BulkWritableCircuitRequestTenant{Id: &vmInterfaceStatusId}
	vmInterfaceNullableTenant := nb.NullableBulkWritableCircuitRequestTenant{}
	vmInterfaceNullableTenant.Set(&vmInterfaceTenant)

	ipToInterfaceRequest := nb.IPAddressToInterfaceRequest{
		IpAddress:   nb.BulkWritableCableRequestStatus{Id: &ipAddressStatusId},
		VmInterface: vmInterfaceNullableTenant,
	}

	_, _, err := c.IpamAPI.IpamIpAddressToInterfaceCreate(auth).IPAddressToInterfaceRequest(ipToInterfaceRequest).Execute()
	if err != nil {
		return fmt.Errorf("failed to assign IP address to VM interface: %v", err)
	}
	return nil
}

func removeIPAddressFromVMInterface(ctx context.Context, c *nb.APIClient, token, ipAddressID, vmInterfaceID string) error {
	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {Key: token, Prefix: "Token"},
		},
	)

	ipAddress, _, err := c.IpamAPI.IpamIpAddressesRetrieve(auth, ipAddressID).Execute()
	if err != nil {
		return fmt.Errorf("failed to retrieve IP address: %v", err)
	}

	var assignmentID string
	for _, vmInterface := range ipAddress.VmInterfaces {
		if vmInterface.Id != nil && vmInterface.Id.String != nil && *vmInterface.Id.String == vmInterfaceID {
			assignmentID = *vmInterface.Id.String
			break
		}
	}

	if assignmentID == "" {
		return fmt.Errorf("no assignment found for IP address %s and VM interface %s", ipAddressID, vmInterfaceID)
	}

	_, err = c.IpamAPI.IpamIpAddressToInterfaceDestroy(auth, assignmentID).Execute()
	if err != nil {
		return fmt.Errorf("failed to remove IP address assignment: %v", err)
	}
	return nil
}
