package nodes

import (
	"github.com/catalisaio/workflow-core/pkg/types"
)

// SplitInBatchesNode implements the Split In Batches node.
type SplitInBatchesNode struct{}

// Type returns the n8n node type.
func (n *SplitInBatchesNode) Type() string {
	return "n8n-nodes-base.splitInBatches"
}

// Execute splits input items into batches.
func (n *SplitInBatchesNode) Execute(ctx *types.ExecutionContext, params map[string]interface{}, input []types.NodeData) ([]types.NodeData, error) {
	if len(input) == 0 {
		return types.EmptyNodeData(), nil
	}

	batchSize := getIntParam(params, "batchSize", 10)
	if batchSize <= 0 {
		batchSize = 10
	}

	// Get options
	options := getMapParam(params, "options")
	reset := false
	if options != nil {
		reset = getBoolParam(options, "reset", false)
	}

	// Check if we have batch state
	batchIndex, hasBatchState := ctx.GetVariable("splitInBatches_index")
	if !hasBatchState || reset {
		batchIndex = 0
		ctx.SetVariable("splitInBatches_items", input)
	}

	idx, ok := batchIndex.(int)
	if !ok {
		idx = 0
	}

	// Get stored items
	storedItems, _ := ctx.GetVariable("splitInBatches_items")
	items, ok := storedItems.([]types.NodeData)
	if !ok {
		items = input
	}

	// Calculate batch
	start := idx * batchSize
	end := start + batchSize
	if end > len(items) {
		end = len(items)
	}

	if start >= len(items) {
		// No more items, signal completion
		return []types.NodeData{}, nil
	}

	// Get batch
	batch := items[start:end]

	// Store next batch index
	ctx.SetVariable("splitInBatches_index", idx+1)

	// Add batch metadata
	for i := range batch {
		if batch[i].JSON == nil {
			batch[i].JSON = make(map[string]interface{})
		}
		batch[i].JSON["$batch"] = map[string]interface{}{
			"index":     idx,
			"size":      len(batch),
			"total":     len(items),
			"remaining": len(items) - end,
		}
	}

	return batch, nil
}
