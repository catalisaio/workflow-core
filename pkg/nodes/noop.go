package nodes

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// NoOpNode implements the No Operation node (passthrough).
type NoOpNode struct{}

// Type returns the n8n node type.
func (n *NoOpNode) Type() string {
	return "n8n-nodes-base.noOp"
}

// Execute passes through input data unchanged.
func (n *NoOpNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}
	return input, nil
}
