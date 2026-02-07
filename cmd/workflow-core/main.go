// Package main provides the CLI entrypoint for workflow-core.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/catalisaio/workflow-core/pkg/engine"
	"github.com/catalisaio/workflow-core/pkg/nodes"
	"github.com/catalisaio/workflow-core/pkg/parser"
)

var (
	version = "0.1.0"
	commit  = "dev"
)

func main() {
	// Define subcommands
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runInputFile := runCmd.String("input", "", "JSON file with input data")
	runTimeout := runCmd.Duration("timeout", 30*time.Minute, "Execution timeout")
	runVerbose := runCmd.Bool("verbose", false, "Verbose output")

	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)

	listNodesCmd := flag.NewFlagSet("list-nodes", flag.ExitOnError)

	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
		if runCmd.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "Error: workflow file required")
			fmt.Fprintln(os.Stderr, "Usage: workflow-core run [options] <workflow.json>")
			os.Exit(1)
		}
		workflowFile := runCmd.Arg(0)
		exitCode := runWorkflow(workflowFile, *runInputFile, *runTimeout, *runVerbose)
		os.Exit(exitCode)

	case "validate":
		validateCmd.Parse(os.Args[2:])
		if validateCmd.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "Error: workflow file required")
			os.Exit(1)
		}
		workflowFile := validateCmd.Arg(0)
		exitCode := validateWorkflow(workflowFile)
		os.Exit(exitCode)

	case "list-nodes":
		listNodesCmd.Parse(os.Args[2:])
		listSupportedNodes()

	case "version":
		versionCmd.Parse(os.Args[2:])
		fmt.Printf("workflow-core version %s (%s)\n", version, commit)

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`workflow-core - High-performance workflow execution engine

Usage:
  workflow-core <command> [options] [arguments]

Commands:
  run         Execute a workflow
  validate    Validate a workflow file
  list-nodes  List supported node types
  version     Show version information
  help        Show this help message

Run Options:
  -input <file>     JSON file with input data
  -timeout <dur>    Execution timeout (default: 30m)
  -verbose          Enable verbose output

Examples:
  workflow-core run workflow.json
  workflow-core run -input data.json -verbose workflow.json
  workflow-core validate workflow.json
  workflow-core list-nodes`)
}

func runWorkflow(workflowFile, inputFile string, timeout time.Duration, verbose bool) int {
	// Parse workflow
	p := parser.NewParser()
	workflow, err := p.ParseFile(workflowFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing workflow: %v\n", err)
		return 1
	}

	if verbose {
		fmt.Printf("Loaded workflow: %s\n", workflow.Name)
		fmt.Printf("Nodes: %d\n", len(workflow.Nodes))
	}

	// Load input data if provided
	var inputData []engine.NodeData
	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			return 1
		}

		var input interface{}
		if err := json.Unmarshal(data, &input); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input JSON: %v\n", err)
			return 1
		}

		// Convert input to NodeData
		switch v := input.(type) {
		case map[string]interface{}:
			inputData = []engine.NodeData{{JSON: v}}
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					inputData = append(inputData, engine.NodeData{JSON: m})
				}
			}
		}
	}

	// Create engine with registry
	opts := engine.DefaultEngineOptions()
	opts.DefaultTimeout = timeout
	eng := engine.NewEngine(opts)
	eng.SetRegistry(nodes.NewRegistry())

	// Set up context with signal handling
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, canceling execution...")
		cancel()
	}()

	// Execute workflow
	if verbose {
		fmt.Println("Starting workflow execution...")
	}

	startTime := time.Now()
	result, err := eng.Execute(ctx, workflow, inputData)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		if result != nil {
			printExecutionResult(result, verbose)
		}
		return 1
	}

	// Print results
	fmt.Printf("\n=== Execution Complete ===\n")
	fmt.Printf("Workflow: %s\n", result.WorkflowName)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Nodes executed: %d\n", len(result.NodeResults))

	printExecutionResult(result, verbose)

	if result.Status == engine.StatusSuccess {
		return 0
	}
	return 1
}

func printExecutionResult(result *engine.ExecutionResult, verbose bool) {
	if verbose {
		fmt.Println("\n--- Node Results ---")
		for name, nodeResult := range result.NodeResults {
			fmt.Printf("\n[%s] %s\n", nodeResult.Status, name)
			fmt.Printf("  Type: %s\n", nodeResult.NodeType)
			fmt.Printf("  Duration: %v\n", nodeResult.Duration)
			if nodeResult.Error != nil {
				fmt.Printf("  Error: %s\n", nodeResult.Error.Message)
			}
		}
	}

	if result.FinalOutput != nil && len(result.FinalOutput) > 0 {
		fmt.Println("\n--- Final Output ---")
		for i, output := range result.FinalOutput {
			jsonData, err := json.MarshalIndent(output.JSON, "", "  ")
			if err != nil {
				fmt.Printf("Item %d: %v\n", i, output.JSON)
			} else {
				fmt.Printf("Item %d:\n%s\n", i, string(jsonData))
			}
		}
	}
}

func validateWorkflow(workflowFile string) int {
	p := parser.NewParser()
	p.SetStrictMode(true)
	
	workflow, err := p.ParseFile(workflowFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		return 1
	}

	// Check for unsupported nodes
	registry := nodes.NewRegistry()
	
	var unsupportedNodes []string
	for _, node := range workflow.Nodes {
		if !registry.Has(node.Type) {
			unsupportedNodes = append(unsupportedNodes, fmt.Sprintf("%s (%s)", node.Name, node.Type))
		}
	}

	fmt.Printf("Workflow: %s\n", workflow.Name)
	fmt.Printf("Nodes: %d\n", len(workflow.Nodes))
	fmt.Printf("Active: %v\n", workflow.Active)

	if len(unsupportedNodes) > 0 {
		fmt.Printf("\nWarning: Unsupported node types:\n")
		for _, node := range unsupportedNodes {
			fmt.Printf("  - %s\n", node)
		}
	}

	fmt.Println("\n✓ Workflow is valid")
	return 0
}

func listSupportedNodes() {
	registry := nodes.NewRegistry()
	nodeTypes := registry.List()

	fmt.Println("Supported Node Types:")
	fmt.Println("=====================")
	
	// Categorize nodes
	categories := map[string][]string{
		"Triggers":    {},
		"Data":        {},
		"Control":     {},
		"HTTP":        {},
		"Other":       {},
	}

	for _, nodeType := range nodeTypes {
		switch {
		case contains(nodeType, "Trigger", "webhook", "cron"):
			categories["Triggers"] = append(categories["Triggers"], nodeType)
		case contains(nodeType, "set", "merge", "split"):
			categories["Data"] = append(categories["Data"], nodeType)
		case contains(nodeType, "if", "switch", "noOp"):
			categories["Control"] = append(categories["Control"], nodeType)
		case contains(nodeType, "http", "respondToWebhook"):
			categories["HTTP"] = append(categories["HTTP"], nodeType)
		default:
			categories["Other"] = append(categories["Other"], nodeType)
		}
	}

	for category, nodeList := range categories {
		if len(nodeList) > 0 {
			fmt.Printf("\n%s:\n", category)
			for _, node := range nodeList {
				fmt.Printf("  - %s\n", node)
			}
		}
	}

	fmt.Printf("\nTotal: %d node types\n", len(nodeTypes))
}

func contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(substr) > 0 && len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
