package models

import "testing"

func TestVehicleType_IsValid(t *testing.T) {
	cases := []struct {
		name  string
		value VehicleType
		want  bool
	}{
		{name: "economy", value: VehicleTypeEconomy, want: true},
		{name: "premium", value: VehicleTypePremium, want: true},
		{name: "xl", value: VehicleTypeXL, want: true},
		{name: "invalid", value: VehicleType("LUX"), want: false},
		{name: "empty", value: VehicleType(""), want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.value.IsValid(); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestPricingTable_ContainsAllVehicleTypes(t *testing.T) {
	if _, ok := PricingTable[VehicleTypeEconomy]; !ok {
		t.Fatal("missing economy pricing")
	}
	if _, ok := PricingTable[VehicleTypePremium]; !ok {
		t.Fatal("missing premium pricing")
	}
	if _, ok := PricingTable[VehicleTypeXL]; !ok {
		t.Fatal("missing xl pricing")
	}
}

func TestPricingTable_ValuesPositive(t *testing.T) {
	for vt, info := range PricingTable {
		if info.BaseFare <= 0 || info.RatePerKm <= 0 || info.RatePerMin <= 0 {
			t.Fatalf("non-positive pricing for %s", vt)
		}
	}
}
