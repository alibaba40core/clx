package yamlutil

// Node is a strict-subset YAML tree (maps, scalar values, and lists).
type Node struct {
	Scalar   string
	Children map[string]*Node
	List     []*Node
}

// GetString returns a scalar at the given path.
func (n *Node) GetString(path ...string) (string, bool) {
	cur := n
	for i, p := range path {
		child, ok := cur.Children[p]
		if !ok {
			return "", false
		}
		if i == len(path)-1 {
			return child.Scalar, child.Scalar != "" || len(child.Children) == 0 && len(child.List) == 0
		}
		cur = child
	}
	return "", false
}

// Has reports whether a path exists.
func (n *Node) Has(path ...string) bool {
	cur := n
	for _, p := range path {
		child, ok := cur.Children[p]
		if !ok {
			return false
		}
		cur = child
	}
	return true
}

// GetChild returns the node at path.
func (n *Node) GetChild(path ...string) (*Node, bool) {
	cur := n
	for _, p := range path {
		child, ok := cur.Children[p]
		if !ok {
			return nil, false
		}
		cur = child
	}
	return cur, true
}

// GetStringList returns scalar list items at path.
func (n *Node) GetStringList(path ...string) ([]string, bool) {
	node, ok := n.GetChild(path...)
	if !ok || node == nil {
		return nil, false
	}
	if len(node.List) == 0 {
		return nil, false
	}
	out := make([]string, 0, len(node.List))
	for _, item := range node.List {
		if item.Scalar != "" {
			out = append(out, item.Scalar)
		}
	}
	return out, len(out) > 0
}

// GetMapKeys returns child keys of a map node at path.
func (n *Node) GetMapKeys(path ...string) ([]string, bool) {
	node, ok := n.GetChild(path...)
	if !ok || node == nil || len(node.Children) == 0 {
		return nil, false
	}
	keys := make([]string, 0, len(node.Children))
	for k := range node.Children {
		keys = append(keys, k)
	}
	return keys, true
}
