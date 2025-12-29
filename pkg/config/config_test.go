package config

import "testing"

func TestLoadUsesEnvironment(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "env-key")
	t.Setenv("OPENAI_BASE_URL", "https://custom-openai/api")
	t.Setenv("OPENAI_MODEL", "gpt-test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got := cfg.GetAPIKey("openai"); got != "env-key" {
		t.Fatalf("expected OPENAI_API_KEY to be set, got %q", got)
	}

	if got := cfg.GetDefaultModel("openai"); got != "gpt-test" {
		t.Fatalf("expected default model to be gpt-test, got %q", got)
	}

	if got := cfg.GetProviderBaseURL("openai"); got != "https://custom-openai/api" {
		t.Fatalf("expected custom base URL, got %q", got)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate should succeed when a key is set, got %v", err)
	}
}

func TestValidateRequiresAPIKey(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation to fail when no API keys are configured")
	}
}
