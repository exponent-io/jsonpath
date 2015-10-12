package jsonpath

type pathNode struct {
	matchOn    interface{} // string, or integer
	childNodes []pathNode
	action     DecodeAction
}

func (n *pathNode) match(path JsonPath) (*pathNode, bool) {
	var node *pathNode = n
	for _, ps := range path {
		found := false
		for i, n := range node.childNodes {
			if n.matchOn == ps {
				node = &node.childNodes[i]
				found = true
				break
			} else if _, ok := ps.(int); ok && n.matchOn == AnyIndex {
				node = &node.childNodes[i]
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return node, true
}

type PathActions struct {
	node pathNode
}

type DecodeAction func(d *Decoder)

// Action specifies an action to call on the Decoder when a particular path is encountered.
func (je *PathActions) Action(action DecodeAction, path ...interface{}) {

	var node *pathNode = &je.node
	for _, ps := range path {
		found := false
		for i, n := range node.childNodes {
			if n.matchOn == ps {
				node = &node.childNodes[i]
				found = true
				break
			}
		}
		if !found {
			node.childNodes = append(node.childNodes, pathNode{matchOn: ps})
			node = &node.childNodes[len(node.childNodes)-1]
		}
	}
	node.action = action
}
