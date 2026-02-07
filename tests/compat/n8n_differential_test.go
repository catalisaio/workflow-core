package compat

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/catalisaio/workflow-core/tests/compat/harness"
)

func TestDifferentialParityAgainstN8N(t *testing.T) {
	if os.Getenv("N8N_REFERENCE_CMD") == "" {
		t.Skip("set N8N_REFERENCE_CMD to enable differential parity checks")
	}

	casesDir := filepath.Join("cases")
	cases, err := harness.LoadCases(casesDir)
	if err != nil {
		t.Fatalf("load cases: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/text" {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("Hello"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"service":"compat"}`))
	}))
	defer server.Close()

	for _, c := range cases {
		c := c
		t.Run(c.ID, func(t *testing.T) {
			if !c.DifferentialEnabled() {
				t.Skip("case disabled for differential parity")
			}

			localEnv := map[string]string{}
			n8nEnv := map[string]string{}
			for k, v := range c.Env {
				localEnv[k] = v
				n8nEnv[k] = v
			}
			if path := localEnv["TEST_HTTP_PATH"]; path != "" {
				localEnv["TEST_HTTP_URL"] = server.URL + path
				n8nEnv["TEST_HTTP_URL"] = toDockerReachableURL(server.URL) + path
			}
			if localEnv["TEST_HTTP_URL"] == "" {
				localEnv["TEST_HTTP_URL"] = server.URL
			}
			if n8nEnv["TEST_HTTP_URL"] == "" {
				n8nEnv["TEST_HTTP_URL"] = toDockerReachableURL(server.URL)
			}

			localExec, localErr := harness.RunWorkflowCore(context.Background(), c, localEnv)
			if localErr != nil {
				t.Fatalf("run workflow-core: %v", localErr)
			}

			n8nExec, n8nErr := harness.RunN8NReference(context.Background(), c, n8nEnv)
			if n8nErr != nil {
				if errors.Is(n8nErr, harness.ErrN8NReferenceNotConfigured) {
					t.Skip("n8n reference runner is not configured")
				}
				t.Fatalf("run n8n reference: %v", n8nErr)
			}

			if cmpErr := harness.CompareOutputs(n8nExec.FinalOutput, localExec.FinalOutput); cmpErr != nil {
				t.Fatalf("parity mismatch: %v", cmpErr)
			}
		})
	}
}

func toDockerReachableURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" {
		u.Host = "host.docker.internal:" + u.Port()
	}
	return u.String()
}
