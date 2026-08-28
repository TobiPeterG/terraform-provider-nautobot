package vlan

import "testing"

func TestValidateVLANRange(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"1-4094", "1,10-20,4094", " 10 - 20 "} {
		if err := validateVLANRange(value); err != nil {
			t.Errorf("%q should be valid: %v", value, err)
		}
	}
	for _, value := range []string{"", "0", "4095", "20-10", "1-2-3", "one"} {
		if err := validateVLANRange(value); err == nil {
			t.Errorf("%q should be invalid", value)
		}
	}
}
