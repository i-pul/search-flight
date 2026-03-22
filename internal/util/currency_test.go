package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		amount   float64
		currency string
		want     string
	}{
		{1500000, "IDR", "Rp1.500.000"},
		{850000, "IDR", "Rp850.000"},
		{850000.53, "IDR", "Rp850.000,53"},
		{1500000, "idr", "Rp1.500.000"}, // case-insensitive
		{0, "IDR", "Rp0"},
		{100.50, "USD", "US$100,50"}, // Indonesian locale uses comma as decimal separator
		{999, "IDR", "Rp999"},
		{-1500000, "IDR", "Rp-1.500.000"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatPrice(tt.amount, tt.currency)
			assert.Equal(t, tt.want, got)
		})
	}
}
