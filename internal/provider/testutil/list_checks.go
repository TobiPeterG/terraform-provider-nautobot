package testutil

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"strconv"
)

func CountAtLeast(dsAddr, listAttr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		raw := rs.Primary.Attributes[listAttr+".#"]
		if raw == "" {
			return fmt.Errorf("%s: %s.# is empty", dsAddr, listAttr)
		}

		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d %s, got %d", dsAddr, min, listAttr, n)
		}
		return nil
	}
}

func FindListIndexByAttr(dsAddr, listAttr, matchField, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes[listAttr+".#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		for i := 0; i < n; i++ {
			k := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if rs.Primary.Attributes[k] == want {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find %s.%s=%q", dsAddr, listAttr, matchField, want)
	}
}

func CheckListItemHasAttrs(dsAddr, listAttr, matchField, matchValue string, want map[string]string, requiredSet []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes[listAttr+".#"]
		if rawN == "" {
			return fmt.Errorf("%s: %s.# is empty", dsAddr, listAttr)
		}
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if rs.Primary.Attributes[k] == matchValue {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find %s with %s=%q", dsAddr, listAttr, matchField, matchValue)
		}

		for field, expected := range want {
			k := fmt.Sprintf("%s.%d.%s", listAttr, idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}

		for _, field := range requiredSet {
			k := fmt.Sprintf("%s.%d.%s", listAttr, idx, field)
			if rs.Primary.Attributes[k] == "" {
				return fmt.Errorf("%s: %s expected to be set, got empty", dsAddr, k)
			}
		}

		return nil
	}
}

func CheckListItemAttrNull(dsAddr, listAttr, matchField, matchValue, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes[listAttr+".#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		for i := 0; i < n; i++ {
			matchKey := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if rs.Primary.Attributes[matchKey] != matchValue {
				continue
			}
			key := fmt.Sprintf("%s.%d.%s", listAttr, i, field)
			if _, exists := rs.Primary.Attributes[key]; exists {
				return fmt.Errorf("%s: %s expected to be null, got %q", dsAddr, key, rs.Primary.Attributes[key])
			}
			return nil
		}

		return fmt.Errorf("%s: expected to find %s with %s=%q", dsAddr, listAttr, matchField, matchValue)
	}
}

func CheckListItemAttrEqualsResourceAttr(dsAddr, listAttr, matchField, matchValue, field, resAddr, resField string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		r, ok := s.RootModule().Resources[resAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resAddr)
		}
		want := r.Primary.Attributes[resField]
		if want == "" {
			return fmt.Errorf("%s.%s is empty", resAddr, resField)
		}

		rawN := ds.Primary.Attributes[listAttr+".#"]
		if rawN == "" {
			return fmt.Errorf("%s: %s.# is empty", dsAddr, listAttr)
		}
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if ds.Primary.Attributes[k] == matchValue {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find %s with %s=%q", dsAddr, listAttr, matchField, matchValue)
		}

		k := fmt.Sprintf("%s.%d.%s", listAttr, idx, field)
		got := ds.Primary.Attributes[k]
		if got != want {
			return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, want, got)
		}
		return nil
	}
}

func CheckVLANInListHasAttrs(dsAddr, vlanName string, want map[string]string) resource.TestCheckFunc {
	return CheckListItemHasAttrs(dsAddr, "vlans", "name", vlanName, want, nil)
}

func CheckVMInListHasAttrs(dsAddr, vmName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "cluster_id", "created", "last_updated", "display", "url", "natural_slug", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "virtual_machines", "name", vmName, want, requiredComputed)
}

func CheckPrefixInListHasAttrs(dsAddr, cidr string, want map[string]string) resource.TestCheckFunc {
	return CheckListItemHasAttrs(dsAddr, "prefixes", "prefix", cidr, want, nil)
}

func CheckManufacturerInListHasAttrs(dsAddr, mName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "manufacturers", "name", mName, want, requiredComputed)
}

func CheckNamespaceInListHasAttrs(dsAddr, namespaceName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "namespaces", "name", namespaceName, want, requiredComputed)
}

func CheckClusterInListHasAttrs(dsAddr, clName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "cluster_type_id", "created", "last_updated", "display", "url", "natural_slug", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "clusters", "name", clName, want, requiredComputed)
}

func CheckClusterTypeInListHasAttrs(dsAddr, ctName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "cluster_types", "name", ctName, want, requiredComputed)
}

func CheckTenantInListHasAttrs(dsAddr, tName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "tenants", "name", tName, want, requiredComputed)
}

func CheckTenantGroupInListHasAttrs(dsAddr, tgName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return CheckListItemHasAttrs(dsAddr, "tenant_groups", "name", tgName, want, requiredComputed)
}
