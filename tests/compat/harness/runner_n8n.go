package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrN8NReferenceNotConfigured = errors.New("N8N_REFERENCE_CMD is not configured")

// RunN8NReference executes a case using an external n8n reference runner.
//
// Expected env var:
//
//	N8N_REFERENCE_CMD='runner --workflow {workflow} --input {input}'
//
// The command must print a JSON array of output items to stdout.
func RunN8NReference(ctx context.Context, c Case, envOverrides map[string]string) (*Execution, error) {
	cmdTemplate := strings.TrimSpace(os.Getenv("N8N_REFERENCE_CMD"))
	if cmdTemplate == "" {
		return nil, ErrN8NReferenceNotConfigured
	}

	tmpDir, err := os.MkdirTemp("", "workflow-core-compat-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	workflowPath := filepath.Join(tmpDir, "workflow.json")
	inputPath := filepath.Join(tmpDir, "input.json")

	workflowBytes, err := json.MarshalIndent(c.Workflow, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal workflow: %w", err)
	}
	if err = os.WriteFile(workflowPath, workflowBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write workflow file: %w", err)
	}

	inputBytes, err := json.MarshalIndent(c.Input, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w", err)
	}
	if err = os.WriteFile(inputPath, inputBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write input file: %w", err)
	}

	cmdStr := strings.ReplaceAll(cmdTemplate, "{workflow}", workflowPath)
	cmdStr = strings.ReplaceAll(cmdStr, "{input}", inputPath)

	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	if repoRoot, rootErr := findRepoRoot(); rootErr == nil {
		cmd.Dir = repoRoot
	}

	env := os.Environ()
	merged := mergeEnv(c.Env, envOverrides)
	for k, v := range merged {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("reference command failed: %w; stderr=%s", err, strings.TrimSpace(stderr.String()))
	}

	var out []map[string]interface{}
	if err = json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("parse reference stdout as JSON array: %w; stdout=%s", err, strings.TrimSpace(stdout.String()))
	}

	return &Execution{FinalOutput: out, Raw: out}, nil
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	candidate := wd
	for {
		if _, statErr := os.Stat(filepath.Join(candidate, "go.mod")); statErr == nil {
			return candidate, nil
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			break
		}
		candidate = parent
	}

	return "", fmt.Errorf("go.mod not found from working directory: %s", wd)
}
