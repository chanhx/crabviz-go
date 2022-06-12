package main

import (
	"go/token"
	"path/filepath"
	"sort"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

type Presenter interface {
	genSVG(graph *Graph) string
}

type Graph struct {
	Tables   []Table
	Clusters []Cluster
	// Edges   []*Edges
}

type Table struct {
	Id       string
	Title    string
	Sections []Node
}

type Node struct {
	Title    string
	SubNodes []Node
}

type Cluster struct {
	Title      string
	nodes      []string
	SubCluster []Cluster
}

func genGraph(fileMembers map[string][]ssa.Member, graph *callgraph.Graph) Graph {
	var tables []Table

	for path, members := range fileMembers {
		sort.Slice(members, func(i int, j int) bool {
			return members[i].Pos() > members[j].Pos()
		})

		nodes := make(map[token.Pos][]Node)
		for _, member := range members {
			node := Node{
				Title: member.Name(),
			}

			key := token.NoPos
			if fn, ok := member.(*ssa.Function); ok {
				parent := fn.Parent()
				if parent != nil {
					key = parent.Pos()
				}

				if subNodes, ok := nodes[fn.Pos()]; ok {
					reverse(subNodes)
					node.SubNodes = subNodes
				}
			}

			nodes[key] = append(nodes[key], node)
		}

		reverse(nodes[token.NoPos])
		tables = append(tables, Table{
			Id:       path,
			Title:    filepath.Base(path),
			Sections: nodes[token.NoPos],
		})
	}

	return Graph{
		Tables: tables,
	}
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
