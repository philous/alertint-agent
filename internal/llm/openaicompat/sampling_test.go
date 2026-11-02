// SPDX-License-Identifier: FSL-1.1-ALv2

package openaicompat_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alertint/alertint-agent/internal/llm"
	"github.com/alertint/alertint-agent/internal/llm/openaicompat"
)

func f64(v float64) *float64 { return &v }

// TestSamplingSent verifies llm.temperature and llm.top_p reach the wire, that
// a client without them sends neither key (the host default stays in effect),
// and that an explicit 0 is sent rather than treated as unset.
func TestSamplingSent(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = nil
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_, _ = w.Write([]byte(chatBody(`{"k":1}`, 1, 1)))
	}))
	defer srv.Close()

	plain := newClient(srv, nil)
	if _, err := plain.Complete(context.Background(), "s", llm.Prompt{Prefix: "p"}, []string{"k"}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"temperature", "top_p"} {
		if _, present := body[key]; present {
			t.Errorf("a client without the option must not send %q", key)
		}
	}

	tuned := newClient(srv, func(c *openaicompat.Config) {
		c.Temperature = f64(1)
		c.TopP = f64(0.95)
	})
	if _, err := tuned.Complete(context.Background(), "s", llm.Prompt{Prefix: "p"}, []string{"k"}); err != nil {
		t.Fatal(err)
	}
	if body["temperature"] != 1.0 || body["top_p"] != 0.95 {
		t.Errorf("temperature/top_p = %v/%v, want 1/0.95", body["temperature"], body["top_p"])
	}

	greedy := newClient(srv, func(c *openaicompat.Config) { c.Temperature = f64(0) })
	if _, err := greedy.Complete(context.Background(), "s", llm.Prompt{Prefix: "p"}, []string{"k"}); err != nil {
		t.Fatal(err)
	}
	if v, present := body["temperature"]; !present || v != 0.0 {
		t.Errorf("temperature = %v (present %v), want an explicit 0", v, present)
	}
}
