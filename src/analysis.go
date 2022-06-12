package main

import (
	"fmt"
	"go/build"
	"go/token"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type Analyzer struct {
	prog      *ssa.Program
	pkgs      []*ssa.Package
	mainPkg   *ssa.Package
	callgraph *callgraph.Graph
}

func (a *Analyzer) Analyze(
	dir string,
	tests bool,
	args []string,
) (fileMembers map[string][]ssa.Member, err error) {
	cfg := &packages.Config{
		Mode:       packages.LoadAllSyntax,
		Tests:      tests,
		Dir:        dir,
		BuildFlags: build.Default.BuildTags,
	}

	initial, err := packages.Load(cfg, args...)
	if err != nil {
		return
	}

	if packages.PrintErrors(initial) > 0 {
		err = fmt.Errorf("packages contain errors")
		return
	}

	// prog, pkgs := ssautil.AllPackages(initial, 0)
	prog, pkgs := ssautil.Packages(initial, 0)
	prog.Build()

	fset := prog.Fset
	graph := static.CallGraph(prog)

	fileMembers = make(map[string][]ssa.Member)

	for _, pkg := range pkgs {
		for _, member := range pkg.Members {
			pos := fset.Position(member.Pos())
			filename := pos.Filename

			if filename == "" {
				continue
			}

			if _, ok := fileMembers[filename]; !ok {
				fileMembers[filename] = []ssa.Member{}
			}

			// `pkg.Members` does not include methods,
			// use `graph.Nodes` to get all functions and methods
			if member.Token() == token.FUNC {
				continue
			}

			fileMembers[filename] = append(fileMembers[filename], member)
		}
	}

	for node := range graph.Nodes {
		if node == nil {
			continue
		}

		pos := fset.Position(node.Pos())
		filename := pos.Filename
		if _, ok := fileMembers[filename]; !ok {
			continue
		}

		fileMembers[filename] = append(fileMembers[filename], node)
	}

	a.prog = prog
	a.pkgs = pkgs
	a.callgraph = graph

	return
}
