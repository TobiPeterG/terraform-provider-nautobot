package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceVLANs() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about all VLANs in Nautobot.",

		ReadContext: dataSourceVLANsRead,

		Schema: map[string]*schema.Schema{
			"vlans": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The UUID of the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"vid": {
							Description: "The ID (VID) of the VLAN.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"name": {
							Description: "The name of the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"description": {
							Description: "Description of the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"vlan_group_id": {
							Description: "The ID of the VLAN group.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "The status of the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tenant_id": {
							Description: "The ID of the tenant associated with the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"role_id": {
							Description: "The ID of the role associated with the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"locations": {
							Description: "The IDs of the locations associated with the VLAN.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tags_ids": {
							Description: "The IDs of the tags associated with the VLAN.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created": {
							Description: "The creation date of the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"last_updated": {
							Description: "The last update date of the VLAN.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVLANsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

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

	// Fetch VLANs list
	rsp, _, err := c.IpamAPI.IpamVlansList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get VLANs list: %s", err.Error())
	}

	results := rsp.Results
	list := make([]map[string]interface{}, 0, len(results))

	// Iterate over the results and map each VLAN to the format expected by Terraform
	for _, vlan := range results {
		// Times -> empty string if missing
		createdStr := ""
		if vlan.Created.IsSet() && vlan.Created.Get() != nil {
			createdStr = vlan.Created.Get().Format(time.RFC3339)
		}
		lastUpdatedStr := ""
		if vlan.LastUpdated.IsSet() && vlan.LastUpdated.Get() != nil {
			lastUpdatedStr = vlan.LastUpdated.Get().Format(time.RFC3339)
		}

		// id -> empty string if missing
		idStr := ""
		if vlan.Id != nil {
			idStr = *vlan.Id
		}

		// description -> empty string if missing
		descStr := ""
		if vlan.Description != nil {
			descStr = *vlan.Description
		}

		itemMap := map[string]interface{}{
			"id":           idStr,
			"vid":          int(vlan.Vid),
			"name":         vlan.Name,
			"description":  descStr,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
		}

		// vlan_group_id -> empty string when missing
		vlanGroupID := ""
		if vlan.VlanGroup.IsSet() {
			if vg := vlan.VlanGroup.Get(); vg != nil && vg.Id != nil && vg.Id.String != nil {
				vlanGroupID = *vg.Id.String
			}
		}
		itemMap["vlan_group_id"] = vlanGroupID

		// status -> resolve name; empty string when missing or lookup failure
		statusName := ""
		if vlan.Status.Id != nil && vlan.Status.Id.String != nil {
			statusID := *vlan.Status.Id.String
			if statusID != "" {
				if name, err := getStatusName(ctx, c, t, statusID); err == nil {
					statusName = name
				}
			}
		}
		itemMap["status"] = statusName

		// tenant_id -> empty string when missing
		tenantID := ""
		if vlan.Tenant.IsSet() {
			if tenant := vlan.Tenant.Get(); tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
				tenantID = *tenant.Id.String
			}
		}
		itemMap["tenant_id"] = tenantID

		// role_id -> empty string when missing
		roleID := ""
		if vlan.Role.IsSet() {
			if role := vlan.Role.Get(); role != nil && role.Id != nil && role.Id.String != nil {
				roleID = *role.Id.String
			}
		}
		itemMap["role_id"] = roleID

		// locations -> always a list; empty when none
		locations := make([]string, 0, len(vlan.Locations))
		for _, location := range vlan.Locations {
			if location.Id != nil && location.Id.String != nil {
				locations = append(locations, *location.Id.String)
			}
		}
		itemMap["locations"] = locations

		// tags_ids -> always a list; empty when none
		tags := make([]string, 0, len(vlan.Tags))
		for _, tag := range vlan.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tags = append(tags, *tag.Id.String)
			}
		}
		itemMap["tags_ids"] = tags

		list = append(list, itemMap)
	}

	// Set the VLANs list in the resource data
	if err := d.Set("vlans", list); err != nil {
		return diag.FromErr(err)
	}

	// Always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
