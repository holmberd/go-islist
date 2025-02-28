package islist

// nodeLevel represent a node's level in a list.
type nodeLevel struct {
	next *Node
	span int
}

// Node represents a node in a list.
type Node struct {
	intervalKey IntervalKey
	levels      []nodeLevel
}

func (n *Node) String() string {
	if n == nil {
		return "nil"
	}
	return n.intervalKey.String()
}

// reset resets the node to its original state.
func (n *Node) reset() *Node {
	if n == nil {
		return n
	}
	n.levels = n.levels[:0] // Reset without deallocating the slice.
	n.intervalKey.Start = 0
	n.intervalKey.End = 0
	n.intervalKey.Key = ""
	return n
}
