package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourcePrefixes() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about all prefixes in Nautobot.",

		ReadContext: dataSourcePrefixesRead,

		Schema: map[string]*schema.Schema{
			"prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The UUID of the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"prefix": {
							Description: "The prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"description": {
							Description: "Description of the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "The status of the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"parent_id": {
							Description: "The ID of the parent of this prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"role_id": {
							Description: "The ID of the role associated with the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tenant_id": {
							Description: "The ID of the tenant associated with the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"rir_id": {
							Description: "The ID of the RIR associated with the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"namespace_id": {
							Description: "The ID of the namespace associated with the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"vlan_id": {
							Description: "The UUID of the VLAN the prefix belongs to.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"created": {
							Description: "The creation date of the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"last_updated": {
							Description: "The last update date of the prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePrefixesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	rsp, _, err := c.IpamAPI.IpamPrefixesList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get prefixes: %s", err.Error())
	}

	results := rsp.Results
	list := make([]map[string]interface{}, 0, len(results))

	for _, prefix := range results {
		// Times -> empty string if missing
		createdStr := ""
		if prefix.Created.IsSet() && prefix.Created.Get() != nil {
			createdStr = prefix.Created.Get().Format(time.RFC3339)
		}
		lastUpdatedStr := ""
		if prefix.LastUpdated.IsSet() && prefix.LastUpdated.Get() != nil {
			lastUpdatedStr = prefix.LastUpdated.Get().Format(time.RFC3339)
		}

		// id -> empty string if missing
		idStr := ""
		if prefix.Id != nil {
			idStr = *prefix.Id
		}

		// description -> empty string if missing
		descStr := ""
		if prefix.Description != nil {
			descStr = *prefix.Description
		}

		itemMap := map[string]interface{}{
			"id":           idStr,
			"prefix":       prefix.Prefix, // string (not nullable in spec)
			"description":  descStr,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
		}

		// status -> resolve to name; empty string when missing/unresolved
		statusName := ""
		if prefix.Status.Id != nil && prefix.Status.Id.String != nil {
			if statusID := *prefix.Status.Id.String; statusID != "" {
				if name, err := getStatusName(ctx, c, t, statusID); err == nil {
					statusName = name
				}
			}
		}
		itemMap["status"] = statusName

		// parent_id -> empty string when missing
		parentID := ""
		if prefix.Parent.IsSet() {
			if parent := prefix.Parent.Get(); parent != nil && parent.Id != nil && parent.Id.String != nil {
				parentID = *parent.Id.String
			}
		}
		itemMap["parent_id"] = parentID

		// tenant_id -> empty string when missing
		tenantID := ""
		if prefix.Tenant.IsSet() {
			if tenant := prefix.Tenant.Get(); tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
				tenantID = *tenant.Id.String
			}
		}
		itemMap["tenant_id"] = tenantID

		// role_id -> empty string when missing
		roleID := ""
		if prefix.Role.IsSet() {
			if role := prefix.Role.Get(); role != nil && role.Id != nil && role.Id.String != nil {
				roleID = *role.Id.String
			}
		}
		itemMap["role_id"] = roleID

		// rir_id -> empty string when missing
		rirID := ""
		if prefix.Rir.IsSet() {
			if rir := prefix.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
				rirID = *rir.Id.String
			}
		}
		itemMap["rir_id"] = rirID

		// namespace_id -> empty string when missing
		namespaceID := ""
		if prefix.Namespace != nil && prefix.Namespace.Id != nil && prefix.Namespace.Id.String != nil {
			namespaceID = *prefix.Namespace.Id.String
		}
		itemMap["namespace_id"] = namespaceID

		// vlan_id -> empty string when missing
		vlanID := ""
		if prefix.Vlan.IsSet() {
			if vlan := prefix.Vlan.Get(); vlan != nil && vlan.Id != nil && vlan.Id.String != nil {
				vlanID = *vlan.Id.String
			}
		}
		itemMap["vlan_id"] = vlanID

		list = append(list, itemMap)
	}

	if err := d.Set("prefixes", list); err != nil {
		return diag.FromErr(err)
	}

	// Set ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
