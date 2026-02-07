package model

// ConnectionInfo provides connection details for execution.
type ConnectionInfo struct {
	SourceNode      string
	SourceOutput    string
	SourceIndex     int
	TargetNode      string
	TargetInput     string
	TargetIndex     int
}

// BuildConnectionList extracts all connections as a flat list.
func (c Connections) BuildConnectionList() []ConnectionInfo {
	var list []ConnectionInfo
	
	for sourceName, outputs := range c {
		for outputType, outputGroups := range outputs {
			for outputIdx, targets := range outputGroups {
				for _, target := range targets {
					list = append(list, ConnectionInfo{
						SourceNode:   sourceName,
						SourceOutput: outputType,
						SourceIndex:  outputIdx,
						TargetNode:   target.Node,
						TargetInput:  target.Type,
						TargetIndex:  target.Index,
					})
				}
			}
		}
	}
	
	return list
}

// GetSourceNodes returns all source node names.
func (c Connections) GetSourceNodes() []string {
	var nodes []string
	for nodeName := range c {
		nodes = append(nodes, nodeName)
	}
	return nodes
}

// GetTargetsForNode returns all targets for a given source node.
func (c Connections) GetTargetsForNode(nodeName string) []ConnectionTarget {
	var targets []ConnectionTarget
	
	outputs, ok := c[nodeName]
	if !ok {
		return targets
	}
	
	for _, outputGroups := range outputs {
		for _, group := range outputGroups {
			targets = append(targets, group...)
		}
	}
	
	return targets
}

// HasConnection checks if a connection exists between two nodes.
func (c Connections) HasConnection(sourceNode, targetNode string) bool {
	outputs, ok := c[sourceNode]
	if !ok {
		return false
	}
	
	for _, outputGroups := range outputs {
		for _, targets := range outputGroups {
			for _, t := range targets {
				if t.Node == targetNode {
					return true
				}
			}
		}
	}
	
	return false
}

// AddConnection adds a connection between nodes.
func (c Connections) AddConnection(source, outputType string, outputIndex int, target ConnectionTarget) {
	if c[source] == nil {
		c[source] = make(map[string][][]ConnectionTarget)
	}
	
	// Ensure we have enough output groups
	for len(c[source][outputType]) <= outputIndex {
		c[source][outputType] = append(c[source][outputType], []ConnectionTarget{})
	}
	
	c[source][outputType][outputIndex] = append(c[source][outputType][outputIndex], target)
}
