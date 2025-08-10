package config

import "testing"

// Covers validation branch: LLMAI enabled but no default provider.
func TestConfig_Validate_LLMAIEnabledNoProvider(t *testing.T) {
	c := Default()
	c.LLMAI.Enabled = true
	c.LLMAI.DefaultProvider = "" // force error
	if err := c.Validate(); err == nil {
		t.Fatalf("expected validation error when LLMAI enabled without provider")
	}
}

// Covers GetLLMProviderConfig negative lookup and HasValidLLMProvider false path.
func TestConfig_LLMProviderLookupFailures(t *testing.T) {
	c := Default()
	c.LLMAI.Enabled = true
	if _, ok := c.GetLLMProviderConfig("nonexistent"); ok {
		t.Fatalf("expected nonexistent provider lookup to be false")
	}
	if c.HasValidLLMProvider("nonexistent") {
		t.Fatalf("expected HasValidLLMProvider false for missing provider")
	}
	// Non-enabled path should short-circuit
	c.LLMAI.Enabled = false
	if c.HasValidLLMProvider("deepseek") { // even though deepseek special case, disabled overrides
		t.Fatalf("expected false when LLMAI disabled regardless of provider")
	}
}
