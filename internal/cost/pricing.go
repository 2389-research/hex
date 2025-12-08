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
// Prices are per 1M tokens (divide by 1,000,000 to get per-token cost)
var modelPricing = map[string]PricingModel{
	// Anthropic Claude models (with prompt caching)
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

	// OpenAI models (no prompt caching)
	"gpt-4o": {
		InputTokenPrice:  2.50,
		OutputTokenPrice: 10.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"gpt-4o-mini": {
		InputTokenPrice:  0.15,
		OutputTokenPrice: 0.60,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"o3": {
		InputTokenPrice:  15.00, // Reasoning models are more expensive
		OutputTokenPrice: 60.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"o4-mini": {
		InputTokenPrice:  1.00,
		OutputTokenPrice: 4.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},

	// Google Gemini models (no prompt caching in pricing)
	"gemini-3-pro-preview": {
		InputTokenPrice:  1.25,
		OutputTokenPrice: 5.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"gemini-2.5-flash": {
		InputTokenPrice:  0.075,
		OutputTokenPrice: 0.30,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"gemini-2.5-flash-lite": {
		InputTokenPrice:  0.025,
		OutputTokenPrice: 0.10,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
	},
	"gemini-2.5-pro": {
		InputTokenPrice:  1.25,
		OutputTokenPrice: 5.00,
		CacheReadPrice:   0.00,
		CacheWritePrice:  0.00,
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
