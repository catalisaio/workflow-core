package nodes

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// MergeNode implements the Merge node for combining data from multiple inputs.
type MergeNode struct{}

// Type returns the n8n node type.
func (n *MergeNode) Type() string {
	return "n8n-nodes-base.merge"
}

// Execute merges input data based on mode.
func (n *MergeNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}

	mode := getStringParam(params, "mode", "append")

	switch mode {
	case "append":
		return n.mergeAppend(input), nil
	case "combine":
		return n.mergeCombine(params, input), nil
	case "chooseBranch":
		return n.mergeChooseBranch(params, input), nil
	case "multiplex":
		return n.mergeMultiplex(input), nil
	default:
		return n.mergeAppend(input), nil
	}
}

// mergeAppend appends all items from all inputs.
func (n *MergeNode) mergeAppend(input []types.NodeData) []types.NodeData {
	return input
}

// mergeCombine combines items based on mode.
func (n *MergeNode) mergeCombine(params map[string]interface{}, input []types.NodeData) []types.NodeData {
	combineMode := getStringParam(params, "combineMode", "mergeByPosition")

	switch combineMode {
	case "mergeByPosition":
		return n.mergeByPosition(input)
	case "mergeByFields":
		return n.mergeByFields(params, input)
	case "multiplex":
		return n.mergeMultiplex(input)
	default:
		return n.mergeByPosition(input)
	}
}

// mergeByPosition merges items at the same position.
func (n *MergeNode) mergeByPosition(input []types.NodeData) []types.NodeData {
	if len(input) == 0 {
		return types.EmptyNodeData()
	}

	// For simplicity, merge all data into one
	merged := make(map[string]interface{})
	for _, item := range input {
		for k, v := range item.JSON {
			merged[k] = v
		}
	}

	return []types.NodeData{{JSON: merged}}
}

// mergeByFields merges items based on matching field values.
func (n *MergeNode) mergeByFields(params map[string]interface{}, input []types.NodeData) []types.NodeData {
	// Get the merge fields
	fieldsToMatch := getSliceParam(params, "mergeByFields")
	if fieldsToMatch == nil || len(fieldsToMatch) == 0 {
		return n.mergeByPosition(input)
	}

	// For now, simple implementation
	return n.mergeByPosition(input)
}

// mergeChooseBranch selects items from a specific branch.
func (n *MergeNode) mergeChooseBranch(params map[string]interface{}, input []types.NodeData) []types.NodeData {
	// output := getStringParam(params, "output", "input1")

	// In this simple implementation, just return all input
	// A full implementation would track which input came from which branch
	return input
}

// mergeMultiplex creates all combinations of input items.
func (n *MergeNode) mergeMultiplex(input []types.NodeData) []types.NodeData {
	if len(input) <= 1 {
		return input
	}

	// For simplicity, merge all data
	return n.mergeByPosition(input)
}
