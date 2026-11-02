// SPDX-License-Identifier: FSL-1.1-ALv2

package config

import (
	"strings"
	"testing"
)

func TestLLMSamplingConfig(t *testing.T) {
	cfg := Defaults()
	if cfg.LLM.Temperature != nil || cfg.LLM.TopP != nil {
		t.Fatal("temperature and top_p must default to unset (the host default applies)")
	}

	t.Run("valid", func(t *testing.T) {
		cfg, err := loadFrom(strings.NewReader(gatewayLLMYAML+"  temperature: 1.0\n  top_p: 0.95\n"), "", true)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.LLM.Temperature == nil || *cfg.LLM.Temperature != 1 || cfg.LLM.TopP == nil || *cfg.LLM.TopP != 0.95 {
			t.Fatalf("temperature/top_p = %v/%v", cfg.LLM.Temperature, cfg.LLM.TopP)
		}
	})

	t.Run("explicit zero temperature is kept", func(t *testing.T) {
		cfg, err := loadFrom(strings.NewReader(gatewayLLMYAML+"  temperature: 0\n"), "", true)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.LLM.Temperature == nil || *cfg.LLM.Temperature != 0 {
			t.Fatalf("temperature = %v, want an explicit 0", cfg.LLM.Temperature)
		}
	})

	for _, tc := range []struct{ name, extra, want string }{
		{"negative temperature", "  temperature: -0.1\n", "llm.temperature (-0.1) must be between 0 and 2"},
		{"temperature too high", "  temperature: 2.5\n", "llm.temperature (2.5) must be between 0 and 2"},
		{"zero top_p", "  top_p: 0\n", "llm.top_p (0) must be above 0 and at most 1"},
		{"top_p above 1", "  top_p: 1.5\n", "llm.top_p (1.5) must be above 0 and at most 1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadFrom(strings.NewReader(gatewayLLMYAML+tc.extra), "", true)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("load = %v, want error containing %q", err, tc.want)
			}
		})
	}

	for _, tc := range []struct{ name, extra, want string }{
		{"temperature", "  temperature: 1.0\n", `llm.temperature is only valid for provider "openai-compatible"`},
		{"top_p", "  top_p: 0.9\n", `llm.top_p is only valid for provider "openai-compatible"`},
	} {
		t.Run("anthropic rejects "+tc.name, func(t *testing.T) {
			body := strings.Replace(minimalValidYAML, "llm:\n", "llm:\n"+tc.extra, 1)
			_, err := Load(writeConfig(t, body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load = %v, want error containing %q", err, tc.want)
			}
		})
	}
}
