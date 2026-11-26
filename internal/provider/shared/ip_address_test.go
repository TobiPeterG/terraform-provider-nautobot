package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	nb "github.com/nautobot/go-nautobot/v3"
)

func TestNullableNATInsideHelpers(t *testing.T) {
	t.Parallel()

	var unset nb.NullableNATInside
	if got := NullableNATInsideID(unset).ValueString(); got != "" {
		t.Fatalf("expected unset NATInside to become empty, got %q", got)
	}

	emptyJSON, err := json.Marshal(NullableNATInside(""))
	if err != nil {
		t.Fatalf("marshal empty NATInside: %v", err)
	}
	if string(emptyJSON) != "null" {
		t.Fatalf("expected empty NATInside to marshal as null, got %s", emptyJSON)
	}

	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	if got := NullableNATInsideID(NullableNATInside(want)).ValueString(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNullableIPParentHelpers(t *testing.T) {
	t.Parallel()

	var unset nb.NullableBulkWritableIPAddressRequestParent
	if got := NullableIPParentID(unset).ValueString(); got != "" {
		t.Fatalf("expected unset parent to become empty, got %q", got)
	}

	emptyJSON, err := json.Marshal(NullableIPParent(""))
	if err != nil {
		t.Fatalf("marshal empty parent: %v", err)
	}
	if string(emptyJSON) != "null" {
		t.Fatalf("expected empty parent to marshal as null, got %s", emptyJSON)
	}

	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	if got := NullableIPParentID(NullableIPParent(want)).ValueString(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestIPNamespaceReference(t *testing.T) {
	t.Parallel()

	if got := IPNamespaceReference(""); got != nil {
		t.Fatalf("expected nil namespace for empty id, got %+v", got)
	}
	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	ns := IPNamespaceReference(want)
	if ns == nil || ns.Id == nil || ns.Id.String == nil || *ns.Id.String != want {
		t.Fatalf("expected namespace id %q, got %+v", want, ns)
	}
}

func TestIPAddressTypeHelpers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want nb.IPAddressTypeChoices
	}{
		{"", nb.IPADDRESSTYPECHOICES_HOST},
		{"host", nb.IPADDRESSTYPECHOICES_HOST},
		{"DHCP", nb.IPADDRESSTYPECHOICES_DHCP},
		{"slaac", nb.IPADDRESSTYPECHOICES_SLAAC},
	} {
		got, err := IPAddressTypeFromString(tc.in)
		if err != nil {
			t.Fatalf("IPAddressTypeFromString(%q): unexpected err: %v", tc.in, err)
		}
		if got == nil || *got != tc.want {
			t.Fatalf("IPAddressTypeFromString(%q): want %q, got %v", tc.in, tc.want, got)
		}
		if s := IPAddressTypeValue(got).ValueString(); s != string(tc.want) {
			t.Fatalf("IPAddressTypeValue(%q): want %q, got %q", tc.want, tc.want, s)
		}
	}

	if _, err := IPAddressTypeFromString("invalid"); err == nil {
		t.Fatalf("expected error for invalid IP address type")
	}

	if got := IPAddressTypeValue(nil).ValueString(); got != string(nb.IPADDRESSTYPECHOICES_HOST) {
		t.Fatalf("expected nil IP type to default to host, got %q", got)
	}
}

func TestExpandTagSetRefsAndReferenceIDSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id1 := "11111111-1111-1111-1111-111111111111"
	id2 := "22222222-2222-2222-2222-222222222222"

	in := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue(id2),
		types.StringValue(""),
		types.StringValue(id1),
	})

	refs, diags := ExpandTagSetRefs(ctx, in)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(refs) != 2 || *refs[0].Id.String != id1 || *refs[1].Id.String != id2 {
		t.Fatalf("expected sorted [%s %s], got %+v", id1, id2, refs)
	}

	reversedAPI := []nb.BulkWritableCableRequestStatus{refs[1], refs[0]}
	out := ReferenceIDSet(reversedAPI)
	expectedSet := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue(id1),
		types.StringValue(id2),
	})
	if !out.Equal(expectedSet) {
		t.Fatalf("expected tag set %v to equal %v", out, expectedSet)
	}

	emptyRefs, diags := ExpandTagSetRefs(ctx, types.SetValueMust(types.StringType, []attr.Value{}))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if emptyRefs == nil || len(emptyRefs) != 0 {
		t.Fatalf("expected non-nil empty tag slice, got %#v", emptyRefs)
	}
}

func TestResolveIPAddressNamespaceID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	wantNS := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	parentID := "11111111-2222-3333-4444-555555555555"

	ipWithProps := &nb.IPAddress{}
	ipWithProps.Parent.Set(&nb.BulkWritableIPAddressRequestParent{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: StringPtr(parentID)},
		AdditionalProperties: map[string]any{
			"namespace": map[string]any{"id": wantNS},
		},
	})

	got, err := ResolveIPAddressNamespaceID(ctx, nil, ipWithProps)
	if err != nil {
		t.Fatalf("unexpected error from AdditionalProperties fast path: %v", err)
	}
	if got != wantNS {
		t.Fatalf("expected %q, got %q", wantNS, got)
	}

	if _, err := ResolveIPAddressNamespaceID(ctx, nil, &nb.IPAddress{}); err == nil {
		t.Fatalf("expected error when IP address has no parent")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Not found."}`))
	}))
	defer srv.Close()

	cfg := nb.NewConfiguration()
	cfg.Servers = nb.ServerConfigurations{{URL: srv.URL}}
	client := nb.NewAPIClient(cfg)

	ipWithoutProps := &nb.IPAddress{}
	ipWithoutProps.Parent = NullableIPParent(parentID)

	if _, err := ResolveIPAddressNamespaceID(ctx, client, ipWithoutProps); err == nil {
		t.Fatalf("expected error when parent prefix returns 404")
	}
}

func TestRequiresReplaceIfIPHostChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mod := RequiresReplaceIfIPHostChanged()
	nonNullObj := tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})

	cases := []struct {
		name        string
		state, plan string
		wantReplace bool
	}{
		{name: "same cidr", state: "10.0.0.2/28", plan: "10.0.0.2/28", wantReplace: false},
		{name: "mask change only", state: "10.0.0.2/28", plan: "10.0.0.2/24", wantReplace: false},
		{name: "host change", state: "10.0.0.2/28", plan: "10.0.0.3/28", wantReplace: true},
		{name: "ipv6 equivalent host", state: "2001:0db8::1/64", plan: "2001:db8::1/128", wantReplace: false},
		{name: "ipv6 host change", state: "2001:db8::1/64", plan: "2001:db8::2/64", wantReplace: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.StringRequest{
				State:      tfsdk.State{Raw: nonNullObj},
				Plan:       tfsdk.Plan{Raw: nonNullObj},
				StateValue: types.StringValue(tc.state),
				PlanValue:  types.StringValue(tc.plan),
			}
			var resp planmodifier.StringResponse
			mod.PlanModifyString(ctx, req, &resp)
			if resp.RequiresReplace != tc.wantReplace {
				t.Fatalf("state=%q plan=%q: got RequiresReplace=%v, want %v", tc.state, tc.plan, resp.RequiresReplace, tc.wantReplace)
			}
		})
	}
}
