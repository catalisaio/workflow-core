package nodes

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// CronNode implements the cron/schedule trigger node.
type CronNode struct{}

// Type returns the n8n node type.
func (n *CronNode) Type() string {
	return "n8n-nodes-base.cron"
}

// Execute processes cron trigger - passes through with trigger metadata.
func (n *CronNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	// Cron trigger just starts the workflow with empty data
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	// Add cron metadata
	for i := range input {
		if input[i].JSON == nil {
			input[i].JSON = make(map[string]interface{})
		}
		input[i].JSON["$trigger"] = map[string]interface{}{
			"type": "cron",
		}
	}

	return input, nil
}

// ScheduleTriggerNode implements the schedule trigger node (newer version).
type ScheduleTriggerNode struct{}

// Type returns the n8n node type.
func (n *ScheduleTriggerNode) Type() string {
	return "n8n-nodes-base.scheduleTrigger"
}

// Execute processes schedule trigger.
func (n *ScheduleTriggerNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	// Schedule trigger just starts the workflow with empty data
	if len(input) == 0 {
		input = types.EmptyNodeData()
	}

	// Add schedule metadata
	for i := range input {
		if input[i].JSON == nil {
			input[i].JSON = make(map[string]interface{})
		}
		input[i].JSON["$trigger"] = map[string]interface{}{
			"type": "schedule",
		}
	}

	return input, nil
}
