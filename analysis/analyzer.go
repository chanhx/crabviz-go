package analysis

import (
	"fmt"
	"go/build"
	"go/token"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type Analyzer struct {
	Prog      *ssa.Program
	Pkgs      []*ssa.Package
	MainPkg   *ssa.Package
	Callgraph *callgraph.Graph
}

func (a *Analyzer) Analyze(
	dir string,
	tests bool,
	args []string,
) (fileMembers map[string][]ssa.Member, err error) {
	modulePath, err := getModulePath()
	if err != nil {
		return nil, err
	}

	execPath, err := filepath.Abs("./")
	if err != nil {
		return nil, err
	}

	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedDeps |
		packages.NeedModule

	cfg := &packages.Config{
		Mode:       mode,
		Tests:      tests,
		Dir:        dir,
		BuildFlags: build.Default.BuildTags,
	}

	initial, err := packages.Load(cfg, fmt.Sprintf("%s/...", modulePath))
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

			switch member.Token() {
			// `pkg.Members` does not include methods,
			// use `graph.Nodes` to get all functions and methods
			case token.CONST, token.VAR, token.FUNC:
				continue
			}

			fileMembers[filename] = append(fileMembers[filename], member)
		}
	}

	for fn := range graph.Nodes {
		if fn == nil {
			continue
		}

		pos := fset.Position(fn.Pos())
		filename := pos.Filename

		if filename == "" {
			continue
		}
		if !strings.HasPrefix(filename, execPath) {
			continue
		}

		fileMembers[filename] = append(fileMembers[filename], fn)
	}

	a.Prog = prog
	a.Pkgs = pkgs
	a.Callgraph = graph

	return
}

func getModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed to get module path: %v", err)
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
