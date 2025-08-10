package config

import "testing"

func TestConfig_LLMAIProviderHelpers(t *testing.T) {
	c := Default()
	// Disabled => no providers
	if providers := c.GetAvailableLLMProviders(); len(providers) != 0 {
		t.Errorf("expected no providers when disabled")
	}

	c.LLMAI.Enabled = true
	// Ensure deepseek considered valid even without key
	if !c.HasValidLLMProvider("deepseek") {
		t.Errorf("expected deepseek valid without key")
	}
	if _, ok := c.GetLLMProviderConfig("openai"); !ok {
		t.Errorf("expected openai provider present")
	}
	provs := c.GetAvailableLLMProviders()
	found := false
	for _, p := range provs {
		if p == "deepseek" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected deepseek in available providers: %v", provs)
	}
}
