package config

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/currency"
)

const DefaultCurrency = "USD"

func CurrencyFromEnv() (string, error) {
	configuredCurrency := strings.ToUpper(strings.TrimSpace(os.Getenv("CURRENCY")))
	if configuredCurrency == "" {
		return DefaultCurrency, nil
	}
	unit, err := currency.ParseISO(configuredCurrency)
	if err != nil || unit == currency.XXX {
		return "", fmt.Errorf("unsupported CURRENCY %q", configuredCurrency)
	}
	return configuredCurrency, nil
}
