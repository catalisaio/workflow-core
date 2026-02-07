package compat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/catalisaio/workflow-core/tests/compat/harness"
)

func TestWorkflowCoreContractCases(t *testing.T) {
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
			env := map[string]string{}
			for k, v := range c.Env {
				env[k] = v
			}
			if path := env["TEST_HTTP_PATH"]; path != "" {
				env["TEST_HTTP_URL"] = server.URL + path
			}
			if env["TEST_HTTP_URL"] == "" {
				env["TEST_HTTP_URL"] = server.URL
			}

			execResult, runErr := harness.RunWorkflowCore(context.Background(), c, env)
			if runErr != nil {
				t.Fatalf("run workflow-core: %v", runErr)
			}

			if cmpErr := harness.CompareOutputs(c.Expected, execResult.FinalOutput); cmpErr != nil {
				t.Fatalf("%s: %v", c.Description, cmpErr)
			}
		})
	}
}
