package main

import "fmt"

func main() {
	analyzer := new(Analyzer)
	fileMembers, err := analyzer.Analyze("", false, []string{})
	if err != nil {
		panic(err)
	}

	graph := genGraph(fileMembers, analyzer.callgraph)

	dot, err := renderDot(&graph)
	if err != nil {
		panic(err)
	}

	fmt.Println(dot)
}
