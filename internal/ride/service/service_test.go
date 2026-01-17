package service

import "testing"

func TestValidateLanLon(t *testing.T) {
	cases := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{name: "valid", lat: 43.238949, lon: 76.889709, wantErr: false},
		{name: "lat too low", lat: -91, lon: 0, wantErr: true},
		{name: "lat too high", lat: 91, lon: 0, wantErr: true},
		{name: "lon too low", lat: 0, lon: -181, wantErr: true},
		{name: "lon too high", lat: 0, lon: 181, wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateLanLon(tc.lat, tc.lon)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
