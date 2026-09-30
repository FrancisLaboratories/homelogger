package config

import "testing"

func TestCurrencyFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    string
		wantErr bool
	}{
		{name: "missing defaults to USD", want: DefaultCurrency},
		{name: "normalizes currency", env: " eur ", want: "EUR"},
		{name: "accepts common ISO currency", env: "cny", want: "CNY"},
		{name: "rejects unsupported currency", env: "XYZ", wantErr: true},
		{name: "rejects no currency code", env: "XXX", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CURRENCY", tt.env)
			got, err := CurrencyFromEnv()
			if (err != nil) != tt.wantErr {
				t.Fatalf("CurrencyFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CurrencyFromEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}
