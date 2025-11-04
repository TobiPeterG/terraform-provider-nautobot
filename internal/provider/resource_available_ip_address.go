package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceAvailableIPAddress() *schema.Resource {
	return &schema.Resource{
		Description: "This object allocates and manages an available IP address in Nautobot",

		CreateContext: resourceAvailableIPAddressCreate,
		ReadContext:   resourceAvailableIPAddressRead,
		UpdateContext: resourceAvailableIPAddressUpdate,
		DeleteContext: resourceAvailableIPAddressDelete,

		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Description: "ID of the prefix to allocate the IP address from.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"address": {
				Description: "Allocated IP address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ip_version": {
				Description: "IP version of the allocated IP address (4 or 6).",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"dns_name": {
				Description: "DNS name associated with the IP address.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"status": {
				Description: "Status of the allocated IP address.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceAvailableIPAddressCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	prefixID := d.Get("prefix_id").(string)

	statusName := d.Get("status").(string)
	statusID, err := getStatusID(ctx, c, t, statusName)
	if err != nil {
		return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
	}

	ipRequest := nb.IPAllocationRequest{
		Status: nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{String: &statusID},
		},
	}
	if v, ok := d.GetOk("dns_name"); ok && v.(string) != "" {
		dnsName := v.(string)
		ipRequest.DnsName = &dnsName
	}

	rsp, _, err := c.IpamAPI.IpamPrefixesAvailableIpsCreate(auth, prefixID).IPAllocationRequest([]nb.IPAllocationRequest{ipRequest}).Execute()
	if err != nil {
		return diag.Errorf("failed to allocate IP address: %s", err.Error())
	}
	if len(rsp) == 0 {
		return diag.Errorf("no IP address returned from allocation")
	}
	if rsp[0].Id == nil || *rsp[0].Id == "" {
		return diag.Errorf("allocated IP address returned no id")
	}
	d.SetId(*rsp[0].Id)

	d.Set("address", rsp[0].Address)
	d.Set("ip_version", int(rsp[0].IpVersion))
	if rsp[0].DnsName != nil {
		d.Set("dns_name", *rsp[0].DnsName)
	} else {
		d.Set("dns_name", "")
	}

	return resourceAvailableIPAddressRead(ctx, d, meta)
}

func resourceAvailableIPAddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	ipID := d.Id()
	ipAddress, _, err := c.IpamAPI.IpamIpAddressesRetrieve(auth, ipID).Execute()
	if err != nil {
		d.SetId("")
		return diag.Errorf("failed to read IP address %s: %s", ipID, err.Error())
	}

	d.Set("address", ipAddress.Address)
	d.Set("ip_version", int(ipAddress.IpVersion))
	if ipAddress.DnsName != nil {
		d.Set("dns_name", *ipAddress.DnsName)
	} else {
		d.Set("dns_name", "")
	}

	return nil
}

func resourceAvailableIPAddressUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	ipID := d.Id()

	var ipAddress nb.PatchedIPAddressRequest

	if d.HasChange("status") {
		statusName := d.Get("status").(string)
		statusID, err := getStatusID(ctx, c, t, statusName)
		if err != nil {
			return diag.Errorf("failed to get status ID for %s: %s", statusName, err.Error())
		}
		ipAddress.Status = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{String: &statusID},
		}
	}

	if d.HasChange("dns_name") {
		dnsName := d.Get("dns_name").(string)
		if dnsName == "" {
			ipAddress.DnsName = nil
		} else {
			ipAddress.DnsName = &dnsName
		}
	}

	_, _, err := c.IpamAPI.IpamIpAddressesPartialUpdate(auth, ipID).PatchedIPAddressRequest(ipAddress).Execute()
	if err != nil {
		return diag.Errorf("failed to update IP address %s: %s", ipID, err.Error())
	}

	return resourceAvailableIPAddressRead(ctx, d, meta)
}

func resourceAvailableIPAddressDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	auth := context.WithValue(ctx, nb.ContextAPIKeys, map[string]nb.APIKey{
		"tokenAuth": {Key: t, Prefix: "Token"},
	})

	ipID := d.Id()
	_, err := c.IpamAPI.IpamIpAddressesDestroy(auth, ipID).Execute()
	if err != nil {
		return diag.Errorf("failed to delete IP address %s: %s", ipID, err.Error())
	}

	d.SetId("")
	return nil
}
