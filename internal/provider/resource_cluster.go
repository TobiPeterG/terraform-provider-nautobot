package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a cluster in Nautobot",

		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Cluster's name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"comments": {
				Description: "Comments or notes about the cluster.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"cluster_type_id": {
				Description: "ID of the Cluster's type. This can be sourced from the cluster_type resource or data source.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cluster_group_id": {
				Description: "ID of the Cluster's group.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"tenant_id": {
				Description: "ID of the Tenant associated with the cluster.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"location_id": {
				Description: "ID of the Location of the cluster.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"tags_ids": {
				Description: "IDs of the Tags associated with the cluster.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "Creation date of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Last update date of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	clusterName := d.Get("name").(string)
	existingClusters, _, err := c.VirtualizationAPI.VirtualizationClustersList(auth).Name([]string{clusterName}).Execute()
	if err != nil {
		return diag.Errorf("failed to list clusters: %s", err.Error())
	}

	if len(existingClusters.Results) > 0 {
		if existingClusters.Results[0].Id == nil || *existingClusters.Results[0].Id == "" {
			return diag.Errorf("existing cluster %q returned no id", clusterName)
		}
		d.SetId(*existingClusters.Results[0].Id)
		return resourceClusterRead(ctx, d, meta)
	}

	var cluster nb.ClusterRequest
	cluster.Name = clusterName
	cluster.ClusterType = nb.BulkWritableCableRequestStatus{
		Id: &nb.BulkWritableCableRequestStatusId{String: stringPtr(d.Get("cluster_type_id").(string))},
	}

	if v, ok := d.GetOk("comments"); ok && v.(string) != "" {
		comments := v.(string)
		cluster.Comments = &comments
	}
	if v, ok := d.GetOk("cluster_group_id"); ok && v.(string) != "" {
		var cg nb.NullableBulkWritableCircuitRequestTenant
		cg.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{String: stringPtr(v.(string))},
		})
		cluster.ClusterGroup = cg
	}
	if v, ok := d.GetOk("tenant_id"); ok && v.(string) != "" {
		var tenant nb.NullableBulkWritableCircuitRequestTenant
		tenant.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{String: stringPtr(v.(string))},
		})
		cluster.Tenant = tenant
	}
	if v, ok := d.GetOk("location_id"); ok && v.(string) != "" {
		var loc nb.NullableBulkWritableCircuitRequestTenant
		loc.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{String: stringPtr(v.(string))},
		})
		cluster.Location = loc
	}
	if v, ok := d.GetOk("tags_ids"); ok {
		var tags []nb.BulkWritableCableRequestStatus
		for _, tag := range v.([]interface{}) {
			tagStr := tag.(string)
			if tagStr == "" {
				continue
			}
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.BulkWritableCableRequestStatusId{String: stringPtr(tagStr)},
			})
		}
		cluster.Tags = tags
	}

	rsp, _, err := c.VirtualizationAPI.VirtualizationClustersCreate(auth).ClusterRequest(cluster).Execute()
	if err != nil {
		return diag.Errorf("failed to create cluster: %s", err.Error())
	}

	if rsp.Id == nil || *rsp.Id == "" {
		return diag.Errorf("created cluster returned no id")
	}
	d.SetId(*rsp.Id)

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	clusterId := d.Id()
	cluster, _, err := c.VirtualizationAPI.VirtualizationClustersRetrieve(auth, clusterId).Execute()
	if err != nil {
		return diag.Errorf("failed to read cluster: %s", err.Error())
	}

	d.Set("name", cluster.Name)

	if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
		d.Set("cluster_type_id", *cluster.ClusterType.Id.String)
	} else {
		d.Set("cluster_type_id", "")
	}

	if cluster.Comments != nil {
		d.Set("comments", *cluster.Comments)
	} else {
		d.Set("comments", "")
	}

	if cluster.ClusterGroup.IsSet() {
		if clusterGroup := cluster.ClusterGroup.Get(); clusterGroup != nil && clusterGroup.Id != nil && clusterGroup.Id.String != nil {
			d.Set("cluster_group_id", *clusterGroup.Id.String)
		} else {
			d.Set("cluster_group_id", "")
		}
	} else {
		d.Set("cluster_group_id", "")
	}

	if cluster.Tenant.IsSet() {
		if tenant := cluster.Tenant.Get(); tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
			d.Set("tenant_id", *tenant.Id.String)
		} else {
			d.Set("tenant_id", "")
		}
	} else {
		d.Set("tenant_id", "")
	}

	if cluster.Location.IsSet() {
		if location := cluster.Location.Get(); location != nil && location.Id != nil && location.Id.String != nil {
			d.Set("location_id", *location.Id.String)
		} else {
			d.Set("location_id", "")
		}
	} else {
		d.Set("location_id", "")
	}

	tags := make([]string, 0, len(cluster.Tags))
	for _, tag := range cluster.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tags = append(tags, *tag.Id.String)
		}
	}
	d.Set("tags_ids", tags)

	createdStr := ""
	if cluster.Created.IsSet() && cluster.Created.Get() != nil {
		createdStr = cluster.Created.Get().Format(time.RFC3339)
	}
	d.Set("created", createdStr)

	lastUpdatedStr := ""
	if cluster.LastUpdated.IsSet() && cluster.LastUpdated.Get() != nil {
		lastUpdatedStr = cluster.LastUpdated.Get().Format(time.RFC3339)
	}
	d.Set("last_updated", lastUpdatedStr)

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	clusterId := d.Id()

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	var cluster nb.PatchedClusterRequest

	if d.HasChange("name") {
		name := d.Get("name").(string)
		cluster.Name = &name
	}
	if d.HasChange("comments") {
		comments := d.Get("comments").(string)
		if comments == "" {
			cluster.Comments = nil
		} else {
			cluster.Comments = &comments
		}
	}
	if d.HasChange("cluster_type_id") {
		clusterTypeID := d.Get("cluster_type_id").(string)
		cluster.ClusterType = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{String: &clusterTypeID},
		}
	}
	if d.HasChange("cluster_group_id") {
		clusterGroupID := d.Get("cluster_group_id").(string)
		if clusterGroupID == "" {
			cluster.ClusterGroup.Unset()
		} else {
			clusterGroup := &nb.BulkWritableCircuitRequestTenant{
				Id: &nb.BulkWritableCableRequestStatusId{String: &clusterGroupID},
			}
			cluster.ClusterGroup.Set(clusterGroup)
		}
	}
	if d.HasChange("tenant_id") {
		tenantID := d.Get("tenant_id").(string)
		if tenantID == "" {
			cluster.Tenant.Unset()
		} else {
			tenant := &nb.BulkWritableCircuitRequestTenant{
				Id: &nb.BulkWritableCableRequestStatusId{String: &tenantID},
			}
			cluster.Tenant.Set(tenant)
		}
	}
	if d.HasChange("location_id") {
		locationID := d.Get("location_id").(string)
		if locationID == "" {
			cluster.Location.Unset()
		} else {
			location := &nb.BulkWritableCircuitRequestTenant{
				Id: &nb.BulkWritableCableRequestStatusId{String: &locationID},
			}
			cluster.Location.Set(location)
		}
	}
	if d.HasChange("tags_ids") {
		var tags []nb.BulkWritableCableRequestStatus
		for _, tag := range d.Get("tags_ids").([]interface{}) {
			tagID := tag.(string)
			if tagID == "" {
				continue
			}
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.BulkWritableCableRequestStatusId{String: &tagID},
			})
		}
		cluster.Tags = tags
	}

	_, _, err := c.VirtualizationAPI.VirtualizationClustersPartialUpdate(auth, clusterId).PatchedClusterRequest(cluster).Execute()
	if err != nil {
		return diag.Errorf("failed to update cluster: %s", err.Error())
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	clusterId := d.Id()
	_, err := c.VirtualizationAPI.VirtualizationClustersDestroy(auth, clusterId).Execute()
	if err != nil {
		return diag.Errorf("failed to delete cluster: %s", err.Error())
	}

	d.SetId("")
	return nil
}
