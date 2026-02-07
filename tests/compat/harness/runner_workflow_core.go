package harness

import (
	"context"

	"github.com/catalisaio/workflow-core/pkg/engine"
	"github.com/catalisaio/workflow-core/pkg/nodes"
)

// RunWorkflowCore executes one case in workflow-core.
func RunWorkflowCore(ctx context.Context, c Case, envOverrides map[string]string) (*Execution, error) {
	options := engine.DefaultEngineOptions()
	options.Environment = mergeEnv(c.Env, envOverrides)

	eng := engine.NewEngine(options)
	eng.SetRegistry(nodes.NewRegistry())

	input := make([]engine.NodeData, 0, len(c.Input))
	for _, item := range c.Input {
		input = append(input, engine.NodeData{JSON: item})
	}

	result, err := eng.Execute(ctx, &c.Workflow, input)
	if err != nil {
		return nil, err
	}

	output := make([]map[string]interface{}, 0, len(result.FinalOutput))
	for _, item := range result.FinalOutput {
		output = append(output, item.JSON)
	}

	return &Execution{FinalOutput: output, Raw: result}, nil
}

func mergeEnv(base, overrides map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}
