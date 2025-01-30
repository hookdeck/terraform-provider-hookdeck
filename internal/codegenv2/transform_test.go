package codegenv2

import (
	"strings"
	"testing"
)

func TestToConfigCase(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"GitHub", "GitHub"},
		{"HMACAlgorithms", "HmacAlgorithms"},
		{"Shopify", "Shopify"},
		{"TokenIO", "TokenIo"},
		{"AWS SNS", "Awssns"},
		{"3dEye", "3DEye"},
		{"NMI Payment Gateway", "NmiPaymentGateway"},
		{"VerificationHMACConfigs", "VerificationHmacConfigs"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			output := toConfigCase(tc.input)

			if output != tc.expected {
				t.Errorf("toConfig(\"%s\") = \"%s\", expected '%s'", tc.input, output, tc.expected)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"GitHub", "Git Hub"},
		{"HMACAlgorithms", "HMAC Algorithms"},
		{"Shopify", "Shopify"},
		{"TokenIO", "Token IO"},
		{"AWS SNS", "AWS SNS"},
		{"NMI Payment Gateway", "NMI Payment Gateway"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			output := splitWords(tc.input)

			if strings.Join(output, " ") != tc.expected {
				t.Errorf("splitWords(\"%s\") = \"%s\", expected '%s'", tc.input, strings.Join(output, " "), tc.expected)
			}
		})
	}
}
