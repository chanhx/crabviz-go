package graph

import (
	"go/token"
	"hash/fnv"
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
	Edges    []Edge
}

type Table struct {
	Id       uint32
	Title    string
	Sections []Node
}

type Node struct {
	Id       uint32
	Title    string
	SubNodes []Node
	Classes  []string
}

type Edge struct {
	From EdgeNode
	To   EdgeNode
}

type EdgeNode struct {
	TableID uint32
	NodeID  uint32
}

type Cluster struct {
	Title       string
	Nodes       []uint32
	SubClusters []Cluster
}

func GenGraph(fset *token.FileSet, fileMembers map[string][]ssa.Member, graph *callgraph.Graph, pkgFiles map[*ssa.Package][]string) Graph {
	var tables []Table
	edgeSet := make(map[Edge]struct{})

	for path, members := range fileMembers {
		sort.Slice(members, func(i int, j int) bool {
			return members[i].Pos() > members[j].Pos()
		})

		fileID := hash(path)
		nodes := make(map[token.Pos][]Node)

		for _, member := range members {
			nodeID := uint32(member.Pos())

			node := Node{
				Id:    nodeID,
				Title: member.Name(),
			}

			key := token.NoPos
			if fn, ok := member.(*ssa.Function); ok {
				if fn.Pkg != nil {
					node.Title = fn.RelString(fn.Pkg.Pkg)
				}

				node.Id = uint32(graph.Nodes[fn].ID)
				node.Classes = append(node.Classes, "fn")

				parent := fn.Parent()
				if parent != nil {
					key = parent.Pos()
				}

				if subNodes, ok := nodes[fn.Pos()]; ok {
					reverse(subNodes)
					node.SubNodes = subNodes
				}

				for _, edge := range graph.Nodes[fn].In {
					caller := edge.Caller
					callerPos := caller.Func.Pos()
					callerFileID := hash(fset.Position(callerPos).Filename)
					callerID := caller.ID

					edgeSet[Edge{
						EdgeNode{callerFileID, uint32(callerID)},
						EdgeNode{fileID, node.Id},
					}] = struct{}{}
				}
			}

			nodes[key] = append(nodes[key], node)
		}

		reverse(nodes[token.NoPos])
		tables = append(tables, Table{
			Id:       fileID,
			Title:    filepath.Base(path),
			Sections: nodes[token.NoPos],
		})
	}

	edges := make([]Edge, 0, len(edgeSet))
	for edge := range edgeSet {
		edges = append(edges, edge)
	}

	var clusters []Cluster
	for pkg, files := range pkgFiles {
		var nodes []uint32
		for _, file := range files {
			nodes = append(nodes, hash(file))
		}

		if pkg == nil {
			continue
		}

		cluster := Cluster{
			Title: pkg.Pkg.Path(),
			Nodes: nodes,
		}

		clusters = append(clusters, cluster)
	}

	return Graph{
		Tables:   tables,
		Edges:    edges,
		Clusters: clusters,
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
