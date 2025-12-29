package providers

import "testing"

func TestNewProviderReturnsRegisteredProvider(t *testing.T) {
	cfg := ProviderConfig{
		APIKey: "test-key",
		Model:  "gpt-test",
	}

	provider, err := NewProvider("openai", cfg)
	if err != nil {
		t.Fatalf("expected provider to be created, got %v", err)
	}

	if provider.Name() != "openai" {
		t.Fatalf("expected provider name openai, got %s", provider.Name())
	}

	if got := provider.DefaultModel(); got != "gpt-test" {
		t.Fatalf("expected default model to honor config, got %s", got)
	}

	spec, ok := GetProviderSpec("openai")
	if !ok {
		t.Fatal("expected OpenAI spec to be registered")
	}

	if spec.Name != "OpenAI" {
		t.Fatalf("expected spec name OpenAI, got %s", spec.Name)
	}
}

func TestNewProviderUnknown(t *testing.T) {
	if _, err := NewProvider("unknown", ProviderConfig{}); err == nil {
		t.Fatal("expected error when provider is not registered")
	}
}
