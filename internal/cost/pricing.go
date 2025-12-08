// ABOUTME: pricing.go defines per-model pricing for API cost calculations
// ABOUTME: Contains pricing data for Anthropic Claude models

package cost

import (
	"fmt"
)

// PricingModel defines the pricing structure for a model
type PricingModel struct {
	InputTokenPrice  float64 // Per 1M tokens
	OutputTokenPrice float64 // Per 1M tokens
	CacheReadPrice   float64 // Per 1M tokens
	CacheWritePrice  float64 // Per 1M tokens
}

// modelPricing contains pricing for all supported models
var modelPricing = map[string]PricingModel{
	"claude-sonnet-4-5-20250929": {
		InputTokenPrice:  3.00,  // $3.00 per 1M input tokens
		OutputTokenPrice: 15.00, // $15.00 per 1M output tokens
		CacheReadPrice:   0.30,  // $0.30 per 1M cache read tokens
		CacheWritePrice:  3.75,  // $3.75 per 1M cache write tokens
	},
	"claude-sonnet-3-5-20241022": {
		InputTokenPrice:  3.00,
		OutputTokenPrice: 15.00,
		CacheReadPrice:   0.30,
		CacheWritePrice:  3.75,
	},
	"claude-3-5-sonnet-20241022": {
		InputTokenPrice:  3.00,
		OutputTokenPrice: 15.00,
		CacheReadPrice:   0.30,
		CacheWritePrice:  3.75,
	},
	"claude-3-5-sonnet-20240620": {
		InputTokenPrice:  3.00,
		OutputTokenPrice: 15.00,
		CacheReadPrice:   0.30,
		CacheWritePrice:  3.75,
	},
	"claude-3-opus-20240229": {
		InputTokenPrice:  15.00,
		OutputTokenPrice: 75.00,
		CacheReadPrice:   1.50,
		CacheWritePrice:  18.75,
	},
	"claude-3-haiku-20240307": {
		InputTokenPrice:  0.25,
		OutputTokenPrice: 1.25,
		CacheReadPrice:   0.03,
		CacheWritePrice:  0.30,
	},
}

// getPricing returns the pricing model for a given model name
func getPricing(model string) (*PricingModel, error) {
	pricing, exists := modelPricing[model]
	if !exists {
		return nil, fmt.Errorf("unknown model: %s", model)
	}
	return &pricing, nil
}

// calculateTokenCost calculates the cost for a number of tokens
func calculateTokenCost(tokens int64, pricePerMillion float64) float64 {
	return (float64(tokens) / 1_000_000.0) * pricePerMillion
}
