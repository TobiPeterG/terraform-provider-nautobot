package virtual_machine

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type virtualMachineItemModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Status            types.String `tfsdk:"status"`
	TenantID          types.String `tfsdk:"tenant_id"`
	PlatformID        types.String `tfsdk:"platform_id"`
	RoleID            types.String `tfsdk:"role_id"`
	SoftwareVersionID types.String `tfsdk:"software_version_id"`
	PrimaryIP4ID      types.String `tfsdk:"primary_ip4_id"`
	PrimaryIP6ID      types.String `tfsdk:"primary_ip6_id"`
	Vcpus             types.Int64  `tfsdk:"vcpus"`
	Memory            types.Int64  `tfsdk:"memory"`
	Disk              types.Int64  `tfsdk:"disk"`
	Comments          types.String `tfsdk:"comments"`
	TagsIDs           types.List   `tfsdk:"tags_ids"`
	Created           types.String `tfsdk:"created"`
	LastUpdated       types.String `tfsdk:"last_updated"`
	Display           types.String `tfsdk:"display"`
	URL               types.String `tfsdk:"url"`
	NaturalSlug       types.String `tfsdk:"natural_slug"`
	NotesURL          types.String `tfsdk:"notes_url"`
}

func virtualMachineDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":                  dsschema.StringAttribute{Description: "The UUID of the virtual machine.", Computed: true},
		"name":                dsschema.StringAttribute{Description: "The name of the virtual machine.", Computed: true},
		"cluster_id":          dsschema.StringAttribute{Description: "The ID of the cluster associated with the virtual machine.", Computed: true},
		"status":              dsschema.StringAttribute{Description: "The name of the status of the virtual machine.", Computed: true},
		"tenant_id":           dsschema.StringAttribute{Description: "The tenant ID. For name lookup, omit it to match only VMs without a tenant.", Computed: true},
		"platform_id":         dsschema.StringAttribute{Description: "The ID of the platform associated with the virtual machine.", Computed: true},
		"role_id":             dsschema.StringAttribute{Description: "The ID of the role associated with the virtual machine.", Computed: true},
		"software_version_id": dsschema.StringAttribute{Description: "The ID of the installed software version.", Computed: true},
		"primary_ip4_id":      dsschema.StringAttribute{Description: "The ID of the primary IPv4 address.", Computed: true},
		"primary_ip6_id":      dsschema.StringAttribute{Description: "The ID of the primary IPv6 address.", Computed: true},
		"vcpus":               dsschema.Int64Attribute{Description: "The number of virtual CPUs.", Computed: true},
		"memory":              dsschema.Int64Attribute{Description: "The amount of memory in MB.", Computed: true},
		"disk":                dsschema.Int64Attribute{Description: "The disk size in GB.", Computed: true},
		"comments":            dsschema.StringAttribute{Description: "Comments or notes about the virtual machine.", Computed: true},
		"tags_ids":            dsschema.ListAttribute{Description: "The IDs of associated tags.", Computed: true, ElementType: types.StringType},
		"created":             dsschema.StringAttribute{Description: "The creation date (RFC3339).", Computed: true},
		"last_updated":        dsschema.StringAttribute{Description: "The last update date (RFC3339).", Computed: true},
		"display":             dsschema.StringAttribute{Description: "Human-friendly display value for the virtual machine.", Computed: true},
		"url":                 dsschema.StringAttribute{Description: "API URL of the virtual machine.", Computed: true},
		"natural_slug":        dsschema.StringAttribute{Description: "Natural slug for the virtual machine.", Computed: true},
		"notes_url":           dsschema.StringAttribute{Description: "Notes URL for the virtual machine.", Computed: true},
	}
	if selectable {
		for _, name := range []string{"id", "name", "cluster_id", "tenant_id"} {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = true
			attributes[name] = attribute
		}
	}
	return attributes
}

func virtualMachineModelFromAPI(vm *nb.VirtualMachine, statusName func(string) (string, error)) (virtualMachineItemModel, error) {
	if vm == nil || vm.Id == nil || *vm.Id == "" {
		return virtualMachineItemModel{}, fmt.Errorf("virtual machine returned no id")
	}

	status := ""
	if vm.Status.Id != nil && vm.Status.Id.String != nil && *vm.Status.Id.String != "" {
		var err error
		status, err = statusName(*vm.Status.Id.String)
		if err != nil {
			return virtualMachineItemModel{}, fmt.Errorf("resolve virtual machine status: %w", err)
		}
	}

	clusterID := ""
	if vm.Cluster.Id != nil && vm.Cluster.Id.String != nil {
		clusterID = *vm.Cluster.Id.String
	}
	primaryIPv4ID := ""
	if vm.PrimaryIp4.IsSet() {
		if ip := vm.PrimaryIp4.Get(); ip != nil && ip.Id != nil && ip.Id.String != nil {
			primaryIPv4ID = *ip.Id.String
		}
	}
	primaryIPv6ID := ""
	if vm.PrimaryIp6.IsSet() {
		if ip := vm.PrimaryIp6.Get(); ip != nil && ip.Id != nil && ip.Id.String != nil {
			primaryIPv6ID = *ip.Id.String
		}
	}

	return virtualMachineItemModel{
		ID:                types.StringValue(*vm.Id),
		Name:              types.StringValue(vm.Name),
		ClusterID:         types.StringValue(clusterID),
		Status:            types.StringValue(status),
		TenantID:          shared.NullableReferenceID(vm.Tenant),
		PlatformID:        shared.NullableReferenceID(vm.Platform),
		RoleID:            shared.NullableReferenceID(vm.Role),
		SoftwareVersionID: shared.NullableSoftwareVersionID(vm.SoftwareVersion),
		PrimaryIP4ID:      types.StringValue(primaryIPv4ID),
		PrimaryIP6ID:      types.StringValue(primaryIPv6ID),
		Vcpus:             types.Int64Value(nullableInt32Value(vm.Vcpus)),
		Memory:            types.Int64Value(nullableInt32Value(vm.Memory)),
		Disk:              types.Int64Value(nullableInt32Value(vm.Disk)),
		Comments:          types.StringValue(shared.DerefString(vm.Comments)),
		TagsIDs:           shared.ReferenceIDs(vm.Tags),
		Created:           shared.NullableTimeValue(vm.Created),
		LastUpdated:       shared.NullableTimeValue(vm.LastUpdated),
		Display:           types.StringValue(vm.Display),
		URL:               types.StringValue(vm.Url),
		NaturalSlug:       types.StringValue(vm.NaturalSlug),
		NotesURL:          types.StringValue(vm.NotesUrl),
	}, nil
}

func nullableInt32Value(value nb.NullableInt32) int64 {
	if value.IsSet() && value.Get() != nil {
		return int64(*value.Get())
	}
	return 0
}
