package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a specific virtual machine in Nautobot.",

		ReadContext: dataSourceVirtualMachineRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the virtual machine to retrieve.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The UUID of the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_id": {
				Description: "The ID of the cluster associated with the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "The name of the status of the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tenant_id": {
				Description: "The ID of the tenant associated with the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"platform_id": {
				Description: "The ID of the platform associated with the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				Description: "The ID of the role associated with the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"primary_ip4_id": {
				Description: "The ID of the primary IPv4 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"primary_ip6_id": {
				Description: "The ID of the primary IPv6 address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vcpus": {
				Description: "The number of virtual CPUs.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"memory": {
				Description: "The amount of memory in MB.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"disk": {
				Description: "The disk size in GB.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"comments": {
				Description: "Comments or notes about the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tags_ids": {
				Description: "The IDs of the tags associated with the virtual machine.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "The creation date of the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "The last update date of the virtual machine.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Get the virtual machine name from the Terraform configuration
	vmName := d.Get("name").(string)

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    t,
				Prefix: "Token",
			},
		},
	)

	// Fetch virtual machine by name
	rsp, _, err := c.VirtualizationAPI.VirtualizationVirtualMachinesList(auth).Name([]string{vmName}).Execute()
	if err != nil {
		return diag.Errorf("failed to get virtual machine with name %s: %s", vmName, err.Error())
	}

	if len(rsp.Results) == 0 {
		return diag.Errorf("no virtual machine found with name %s", vmName)
	}

	vm := rsp.Results[0]

	// Ensure the ID is present and set it as the Terraform resource ID
	if vm.Id == nil || *vm.Id == "" {
		return diag.Errorf("virtual machine %q returned no id", vmName)
	}
	vmID := *vm.Id
	d.SetId(vmID)

	createdStr := ""
	if vm.Created.IsSet() && vm.Created.Get() != nil {
		createdStr = vm.Created.Get().Format(time.RFC3339)
	}

	lastUpdatedStr := ""
	if vm.LastUpdated.IsSet() && vm.LastUpdated.Get() != nil {
		lastUpdatedStr = vm.LastUpdated.Get().Format(time.RFC3339)
	}

	// Set the fields directly in the resource data
	d.Set("id", vmID)
	d.Set("name", vm.Name)

	// Nullable ints: set to 0 if not present
	if vm.Vcpus.IsSet() && vm.Vcpus.Get() != nil {
		d.Set("vcpus", int(*vm.Vcpus.Get()))
	} else {
		d.Set("vcpus", 0)
	}
	if vm.Memory.IsSet() && vm.Memory.Get() != nil {
		d.Set("memory", int(*vm.Memory.Get()))
	} else {
		d.Set("memory", 0)
	}
	if vm.Disk.IsSet() && vm.Disk.Get() != nil {
		d.Set("disk", int(*vm.Disk.Get()))
	} else {
		d.Set("disk", 0)
	}

	if vm.Comments != nil {
		d.Set("comments", *vm.Comments)
	} else {
		d.Set("comments", "")
	}

	d.Set("created", createdStr)
	d.Set("last_updated", lastUpdatedStr)

	// Extract additional fields (default to "")
	clusterID := ""
	if vm.Cluster.Id != nil && vm.Cluster.Id.String != nil {
		clusterID = *vm.Cluster.Id.String
	}
	d.Set("cluster_id", clusterID)

	statusName := ""
	if vm.Status.Id != nil && vm.Status.Id.String != nil {
		statusID := *vm.Status.Id.String
		if statusID != "" {
			if n, err := getStatusName(ctx, c, t, statusID); err == nil {
				statusName = n
			}
		}
	}
	d.Set("status", statusName)

	tenantID := ""
	if vm.Tenant.IsSet() {
		tenant := vm.Tenant.Get()
		if tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
			tenantID = *tenant.Id.String
		}
	}
	d.Set("tenant_id", tenantID)

	platformID := ""
	if vm.Platform.IsSet() {
		platform := vm.Platform.Get()
		if platform != nil && platform.Id != nil && platform.Id.String != nil {
			platformID = *platform.Id.String
		}
	}
	d.Set("platform_id", platformID)

	roleID := ""
	if vm.Role.IsSet() {
		role := vm.Role.Get()
		if role != nil && role.Id != nil && role.Id.String != nil {
			roleID = *role.Id.String
		}
	}
	d.Set("role_id", roleID)

	primaryIPv4ID := ""
	if vm.PrimaryIp4.IsSet() {
		primaryIp4 := vm.PrimaryIp4.Get()
		if primaryIp4 != nil && primaryIp4.Id != nil && primaryIp4.Id.String != nil {
			primaryIPv4ID = *primaryIp4.Id.String
		}
	}
	d.Set("primary_ip4_id", primaryIPv4ID)

	primaryIPv6ID := ""
	if vm.PrimaryIp6.IsSet() {
		primaryIp6 := vm.PrimaryIp6.Get()
		if primaryIp6 != nil && primaryIp6.Id != nil && primaryIp6.Id.String != nil {
			primaryIPv6ID = *primaryIp6.Id.String
		}
	}
	d.Set("primary_ip6_id", primaryIPv6ID)

	// Handle tags
	tags := make([]interface{}, 0, len(vm.Tags))
	for _, tag := range vm.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tags = append(tags, *tag.Id.String)
		}
	}
	d.Set("tags_ids", tags)

	return diags
}
