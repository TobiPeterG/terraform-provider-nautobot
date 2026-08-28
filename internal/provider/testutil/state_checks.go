package testutil

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func CheckTypeSetContainsResourceAttr(setOwnerAddress, setAttribute, otherAddress, otherAttribute string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		setOwner, ok := state.RootModule().Resources[setOwnerAddress]
		if !ok {
			return fmt.Errorf("not found: %s", setOwnerAddress)
		}
		other, ok := state.RootModule().Resources[otherAddress]
		if !ok {
			return fmt.Errorf("not found: %s", otherAddress)
		}
		want := other.Primary.Attributes[otherAttribute]
		if want == "" {
			return fmt.Errorf("%s.%s is empty", otherAddress, otherAttribute)
		}
		prefix := setAttribute + "."
		for key, value := range setOwner.Primary.Attributes {
			if key != setAttribute+".#" && len(key) >= len(prefix) && key[:len(prefix)] == prefix && value == want {
				return nil
			}
		}
		return fmt.Errorf("%s: expected %s to contain %s.%s=%q", setOwnerAddress, setAttribute, otherAddress, otherAttribute, want)
	}
}

func CheckDataSourceAddressNotEqualAllocated(dataSourceAddress string, allocatedAddresses ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		dataSource, ok := state.RootModule().Resources[dataSourceAddress]
		if !ok {
			return fmt.Errorf("not found: %s", dataSourceAddress)
		}
		value := dataSource.Primary.Attributes["address"]
		if value == "" {
			return fmt.Errorf("%s: address is empty", dataSourceAddress)
		}
		for _, allocated := range allocatedAddresses {
			if allocated != "" && value == allocated {
				return fmt.Errorf("%s: expected data source address to differ from allocated address %q", dataSourceAddress, allocated)
			}
		}
		return nil
	}
}

func CheckDataSourceAddressNotInAllocatedResources(dataSourceAddress string, resourceAddresses ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		dataSource, ok := state.RootModule().Resources[dataSourceAddress]
		if !ok {
			return fmt.Errorf("not found: %s", dataSourceAddress)
		}
		value := dataSource.Primary.Attributes["address"]
		if value == "" {
			return fmt.Errorf("%s: address is empty", dataSourceAddress)
		}
		for _, resourceAddress := range resourceAddresses {
			resourceState, ok := state.RootModule().Resources[resourceAddress]
			if !ok {
				return fmt.Errorf("not found: %s", resourceAddress)
			}
			if allocated := resourceState.Primary.Attributes["address"]; allocated != "" && value == allocated {
				return fmt.Errorf("%s: data source address %q equals allocated %s.address", dataSourceAddress, value, resourceAddress)
			}
		}
		return nil
	}
}
