package islist

import "sync"

// NodePool represents a pool of reusable node objects to use across lists.
// A Pool is safe for concurrent use by multiple goroutines.
type NodePool struct {
	pool sync.Pool
}

func NewNodePool() *NodePool {
	return &NodePool{
		pool: sync.Pool{
			New: func() any {
				return &Node{}
			},
		},
	}
}

// get retrieves a node from the pool or creates a new one.
func (p *NodePool) get() *Node {
	return p.pool.Get().(*Node)
}

// put releases any resources associated with a node and returns it to the pool for reuse.
func (p *NodePool) put(n *Node) {
	p.pool.Put(n.reset())
}
