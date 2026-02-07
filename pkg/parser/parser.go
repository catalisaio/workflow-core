// Package parser provides functionality to parse n8n workflow JSON files.
package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/catalisaio/workflow-core/pkg/model"
)

// Parser handles parsing of n8n workflow files.
type Parser struct {
	strictMode bool
}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{
		strictMode: false,
	}
}

// SetStrictMode enables or disables strict parsing mode.
func (p *Parser) SetStrictMode(strict bool) {
	p.strictMode = strict
}

// ParseFile parses a workflow from a file path.
func (p *Parser) ParseFile(path string) (*model.Workflow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open workflow file: %w", err)
	}
	defer file.Close()

	return p.ParseReader(file)
}

// ParseReader parses a workflow from an io.Reader.
func (p *Parser) ParseReader(r io.Reader) (*model.Workflow, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow data: %w", err)
	}

	return p.Parse(data)
}

// Parse parses a workflow from JSON bytes.
func (p *Parser) Parse(data []byte) (*model.Workflow, error) {
	var workflow model.Workflow

	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow JSON: %w", err)
	}

	if err := p.validate(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Normalize the workflow
	p.normalize(&workflow)

	return &workflow, nil
}

// validate checks the workflow for required fields and consistency.
func (p *Parser) validate(w *model.Workflow) error {
	if len(w.Nodes) == 0 {
		return fmt.Errorf("workflow has no nodes")
	}

	// Check for duplicate node names
	nodeNames := make(map[string]bool)
	for _, node := range w.Nodes {
		if node.Name == "" {
			return fmt.Errorf("node has empty name")
		}
		if nodeNames[node.Name] {
			return fmt.Errorf("duplicate node name: %s", node.Name)
		}
		nodeNames[node.Name] = true
	}

	// Validate connections reference existing nodes
	for sourceName, outputs := range w.Connections {
		if !nodeNames[sourceName] {
			if p.strictMode {
				return fmt.Errorf("connection references non-existent source node: %s", sourceName)
			}
		}
		for _, outputGroups := range outputs {
			for _, targets := range outputGroups {
				for _, target := range targets {
					if !nodeNames[target.Node] {
						if p.strictMode {
							return fmt.Errorf("connection references non-existent target node: %s", target.Node)
						}
					}
				}
			}
		}
	}

	return nil
}

// normalize applies default values and normalizes the workflow structure.
func (p *Parser) normalize(w *model.Workflow) {
	// Ensure connections map is not nil
	if w.Connections == nil {
		w.Connections = make(model.Connections)
	}

	// Normalize nodes
	for i := range w.Nodes {
		if w.Nodes[i].Parameters == nil {
			w.Nodes[i].Parameters = make(map[string]interface{})
		}
		if w.Nodes[i].TypeVersion == 0 {
			w.Nodes[i].TypeVersion = 1
		}
	}
}

// ParseMultiple parses multiple workflows from a slice of JSON data.
func (p *Parser) ParseMultiple(dataList [][]byte) ([]*model.Workflow, error) {
	var workflows []*model.Workflow

	for i, data := range dataList {
		workflow, err := p.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse workflow %d: %w", i, err)
		}
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

// ExtractNodeTypes returns all unique node types used in the workflow.
func ExtractNodeTypes(w *model.Workflow) []string {
	typeSet := make(map[string]bool)
	for _, node := range w.Nodes {
		typeSet[node.Type] = true
	}

	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	return types
}
