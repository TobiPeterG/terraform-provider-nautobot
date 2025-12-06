package provider

import (
	"context"
	"net/http"

	nb "github.com/nautobot/go-nautobot/v3"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	pSchema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nautobotProvider struct{ version string }

type providerModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

type APIClient struct {
	Client *nb.APIClient
	Token  string
}

func New(version string) provider.Provider {
	return &nautobotProvider{version: version}
}

func (p *nautobotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nautobot"
	resp.Version = p.version
}

func (p *nautobotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = pSchema.Schema{
		Attributes: map[string]pSchema.Attribute{
			"url": pSchema.StringAttribute{
				Required:    true,
				Description: "Nautobot API URL",
			},
			"token": pSchema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Admin API token",
			},
		},
	}
}

func (p *nautobotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.URL.IsNull() || cfg.URL.IsUnknown() || cfg.Token.IsNull() || cfg.Token.IsUnknown() {
		resp.Diagnostics.AddError("Missing configuration", "Both url and token must be set")
		return
	}

	conf := nb.NewConfiguration()
	conf.Servers[0].URL = cfg.URL.ValueString()
	api := nb.NewAPIClient(conf)

	api.GetConfig().HTTPClient = &http.Client{
		Transport: &authRT{base: http.DefaultTransport, token: cfg.Token.ValueString()},
	}

	client := &APIClient{Client: api, Token: cfg.Token.ValueString()}
	resp.ResourceData = client
	resp.DataSourceData = client
}

type authRT struct {
	base  http.RoundTripper
	token string
}

func (a *authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Token "+a.token)
	return a.base.RoundTrip(r)
}

func (p *nautobotProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAvailableIPAddressResource,
		NewClusterResource,
		NewClusterTypeResource,
		NewManufacturerResource,
		NewVirtualMachineResource,
		NewVMInterfaceResource,
		NewVMPrimaryIPResource,
	}
}

func (p *nautobotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAvailableIPAddressDataSource,
		NewClusterDataSource,
		NewClustersDataSource,
		NewClusterTypeDataSource,
		NewClusterTypesDataSource,
		NewGraphQLDataSource,
		NewManufacturerDataSource,
		NewManufacturersDataSource,
		NewPrefixDataSource,
		NewPrefixesDataSource,
		NewVirtualMachineDataSource,
		NewVirtualMachinesDataSource,
		NewVLANDataSource,
		NewVLANsDataSource,
	}
}
