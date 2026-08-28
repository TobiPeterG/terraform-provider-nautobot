package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	nb "github.com/nautobot/go-nautobot/v3"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	availableippkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/available_ip_address"
	clusterpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/cluster"
	clustertypepkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/cluster_type"
	graphqlpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/graphql"
	iprangepkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/ip_address_range"
	manufacturerpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/manufacturer"
	namespacepkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/namespace"
	prefixpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/prefix"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
	tenantpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/tenant"
	tenantgrouppkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/tenant_group"
	virtualmachinepkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/virtual_machine"
	vlanpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/vlan"
)

const defaultStatusRequestTimeoutSeconds int64 = 10

type nautobotProvider struct{ version string }

type providerModel struct {
	URL                   types.String `tfsdk:"url"`
	Token                 types.String `tfsdk:"token"`
	SkipVersionCheck      types.Bool   `tfsdk:"skip_version_check"`
	InsecureSkipTLSVerify types.Bool   `tfsdk:"insecure_skip_tls_verify"`
	StatusRequestTimeout  types.Int64  `tfsdk:"status_request_timeout"`
}

func New(version string) provider.Provider {
	return &nautobotProvider{version: version}
}

func (p *nautobotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nautobot"
	resp.Version = p.version
}

func (p *nautobotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		Attributes: map[string]providerschema.Attribute{
			"url": providerschema.StringAttribute{
				Required:    true,
				Description: "Nautobot API URL",
			},
			"token": providerschema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Admin API token",
			},
			"skip_version_check": providerschema.BoolAttribute{
				Optional:    true,
				Description: "Skip Nautobot version compatibility check. Use with caution.",
			},
			"insecure_skip_tls_verify": providerschema.BoolAttribute{
				Optional: true,
				Description: "Disable TLS certificate verification when connecting to Nautobot. " +
					"Use only for testing.",
			},
			"status_request_timeout": providerschema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds for the Nautobot status request used to verify version compatibility. Defaults to 10 seconds. Set to 0 to disable the timeout.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
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

	skipVersionCheck := false
	if !cfg.SkipVersionCheck.IsNull() && !cfg.SkipVersionCheck.IsUnknown() {
		skipVersionCheck = cfg.SkipVersionCheck.ValueBool()
	}

	insecureSkipTLS := false
	if !cfg.InsecureSkipTLSVerify.IsNull() && !cfg.InsecureSkipTLSVerify.IsUnknown() {
		insecureSkipTLS = cfg.InsecureSkipTLSVerify.ValueBool()
	}

	statusRequestTimeoutSeconds := defaultStatusRequestTimeoutSeconds
	if !cfg.StatusRequestTimeout.IsNull() && !cfg.StatusRequestTimeout.IsUnknown() {
		statusRequestTimeoutSeconds = cfg.StatusRequestTimeout.ValueInt64()
	}

	conf := nb.NewConfiguration()
	conf.Servers[0].URL = cfg.URL.ValueString()

	baseTransport := http.DefaultTransport
	if insecureSkipTLS {
		if dt, ok := http.DefaultTransport.(*http.Transport); ok {
			tCopy := dt.Clone()
			tCopy.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			baseTransport = tCopy
		} else {
			baseTransport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
	}

	httpClient := &http.Client{
		Transport: &authRoundTripper{
			base:  baseTransport,
			token: cfg.Token.ValueString(),
		},
	}

	conf.HTTPClient = httpClient
	api := nb.NewAPIClient(conf)

	if !skipVersionCheck {
		if err := p.checkVersionCompatibility(ctx, api, time.Duration(statusRequestTimeoutSeconds)*time.Second); err != nil {
			resp.Diagnostics.AddError(
				"Failed to verify Nautobot version",
				err.Error(),
			)
			return
		}
	}

	client := &shared.APIClient{Client: api}
	resp.ResourceData = client
	resp.DataSourceData = client
}

type authRoundTripper struct {
	base  http.RoundTripper
	token string
}

func (a *authRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	clonedRequest := request.Clone(request.Context())
	clonedRequest.Header = request.Header.Clone()
	clonedRequest.Header.Set("Authorization", "Token "+a.token)
	return a.base.RoundTrip(clonedRequest)
}

func (p *nautobotProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		availableippkg.NewAvailableIPAddressResource,
		clusterpkg.NewClusterResource,
		clustertypepkg.NewClusterTypeResource,
		iprangepkg.NewIPAddressRangeResource,
		manufacturerpkg.NewManufacturerResource,
		namespacepkg.NewNamespaceResource,
		prefixpkg.NewPrefixResource,
		tenantpkg.NewTenantResource,
		tenantgrouppkg.NewTenantGroupResource,
		virtualmachinepkg.NewVirtualMachineResource,
		virtualmachinepkg.NewVMInterfaceResource,
		virtualmachinepkg.NewVMPrimaryIPResource,
		vlanpkg.NewVLANGroupResource,
		vlanpkg.NewVLANResource,
	}
}

func (p *nautobotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		availableippkg.NewAvailableIPAddressDataSource,
		clusterpkg.NewClusterDataSource,
		clusterpkg.NewClustersDataSource,
		clustertypepkg.NewClusterTypeDataSource,
		clustertypepkg.NewClusterTypesDataSource,
		graphqlpkg.NewGraphQLDataSource,
		iprangepkg.NewIPAddressRangeDataSource,
		iprangepkg.NewIPAddressRangesDataSource,
		manufacturerpkg.NewManufacturerDataSource,
		manufacturerpkg.NewManufacturersDataSource,
		namespacepkg.NewNamespaceDataSource,
		namespacepkg.NewNamespacesDataSource,
		prefixpkg.NewPrefixDataSource,
		prefixpkg.NewPrefixesDataSource,
		tenantpkg.NewTenantDataSource,
		tenantpkg.NewTenantsDataSource,
		tenantgrouppkg.NewTenantGroupDataSource,
		tenantgrouppkg.NewTenantGroupsDataSource,
		virtualmachinepkg.NewVirtualMachineDataSource,
		virtualmachinepkg.NewVirtualMachinesDataSource,
		vlanpkg.NewVLANGroupDataSource,
		vlanpkg.NewVLANGroupsDataSource,
		vlanpkg.NewVLANDataSource,
		vlanpkg.NewVLANsDataSource,
	}
}

func (p *nautobotProvider) checkVersionCompatibility(ctx context.Context, api *nb.APIClient, timeout time.Duration) error {
	statusCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		statusCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	status, httpResp, err := api.StatusAPI.StatusRetrieve(statusCtx).Execute()
	if err != nil {
		if timeout > 0 && statusCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("Nautobot /api/status/ request timed out after %s", timeout)
		}
		if httpResp != nil {
			return fmt.Errorf("failed to retrieve Nautobot status: %s (HTTP %s)", err, httpResp.Status)
		}
		return fmt.Errorf("failed to retrieve Nautobot status: %w", err)
	}

	if status == nil || status.NautobotVersion == nil {
		return fmt.Errorf("status endpoint of Nautobot did not return a version")
	}
	nautobotVersion := strings.TrimSpace(*status.NautobotVersion)
	if nautobotVersion == "" {
		return fmt.Errorf("status endpoint of Nautobot did not return a version")
	}

	providerMajMin, err := majorMinorFromVersion(p.version)
	if err != nil {
		return nil
	}
	nautobotMajMin, err := majorMinorFromVersion(nautobotVersion)
	if err != nil {
		return fmt.Errorf("invalid Nautobot version %q: %w", nautobotVersion, err)
	}

	if providerMajMin != nautobotMajMin {
		return fmt.Errorf(
			"provider version %s is only compatible with Nautobot %s.x, but connected instance is %s",
			p.version, providerMajMin, nautobotVersion,
		)
	}

	return nil
}

func majorMinorFromVersion(v string) (string, error) {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		v = v[1:]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("version %q does not have major.minor", v)
	}
	return parts[0] + "." + parts[1], nil
}
