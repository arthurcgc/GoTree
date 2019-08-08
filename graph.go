package main

import "os"

type Node struct {
	info     os.FileInfo
	children []os.FileInfo
	rawFiles []os.FileInfo
	level    int
	hidden   bool
}

type Graph struct {
	nodes []Node
}

func (g *Graph) PushBack(node Node) {
	g.nodes = append(g.nodes, node)
}

func CreateNode(info os.FileInfo, level int, hidden bool) Node {
	node := Node{info, nil, nil, level, hidden}
	return node
}
