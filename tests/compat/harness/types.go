package harness

import "github.com/catalisaio/workflow-core/pkg/model"

// Case describes a single compatibility scenario.
type Case struct {
	ID           string                   `json:"id"`
	Description  string                   `json:"description"`
	Category     string                   `json:"category"`
	Differential *bool                    `json:"differential,omitempty"`
	Workflow     model.Workflow           `json:"workflow"`
	Input        []map[string]interface{} `json:"input"`
	Env          map[string]string        `json:"env,omitempty"`
	Expected     []map[string]interface{} `json:"expected"`
}

func (c Case) DifferentialEnabled() bool {
	if c.Differential == nil {
		return true
	}
	return *c.Differential
}

// Execution captures normalized output and metadata from a runner.
type Execution struct {
	FinalOutput []map[string]interface{}
	Raw         interface{}
}
