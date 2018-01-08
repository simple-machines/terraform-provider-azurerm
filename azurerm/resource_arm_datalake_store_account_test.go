package azurerm

import (
	"testing"
)

func TestValidateArmDataLakeStoreAccountTier(t *testing.T) {
	testCases := []struct {
		input string
		valid bool
	}{
		{"Consumption", true},
		{"Commitment_100TB", true},
		{"consumption", false},
		{"elephant", false},
	}

	for _, test := range testCases {
		_, es := validateArmDataLakeStoreTierType(test.input, "tier")

		if (!test.valid) && len(es) == 0 {
			t.Fatalf("Expected validating tier %q to fail", test.input)
		}
	}
}

func TestValidateArmDataLakeStoreAccountName(t *testing.T) {
	testCases := []struct {
		input string
		valid bool
	}{
		{"a", false},
		{"aaa-123", false},
		{"aaa_123", false},
		{"aaa.123", false},
		{"AAA", false},
		{"aAAA", false},
		{"aaaaaaaaaaaaaaaaaaaaaaaaa", false},
		{"aaa", true},
		{"aaa1", true},
		{"1aaa", true},
		{"aaaaaaaaaaaaaaaaaaaaaaaa", true},
	}

	for _, test := range testCases {
		_, es := validateArmDataLakeStoreAccountName(test.input, "name")

		if (!test.valid) && len(es) == 0 {
			t.Fatalf("Expected validating name %q to fail", test.input)
		}
	}
}
