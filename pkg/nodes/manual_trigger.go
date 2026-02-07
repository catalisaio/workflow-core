package nodes

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// ManualTriggerNode implements the manual trigger node.
type ManualTriggerNode struct{}

// Type returns the n8n node type.
func (n *ManualTriggerNode) Type() string {
	return "n8n-nodes-base.manualTrigger"
}

// Execute runs the manual trigger - it simply passes through any input data.
func (n *ManualTriggerNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	// Manual trigger just passes through input data
	// If no input, return empty data
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}
	return input, nil
}
