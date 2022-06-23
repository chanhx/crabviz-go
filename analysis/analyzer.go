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
	PkgFiles  map[*ssa.Package][]string
	MainPkg   *ssa.Package
	Callgraph *callgraph.Graph
}

func (a *Analyzer) Analyze(
	dir string,
	tests bool,
	args []string,
) (fileMembers map[string][]ssa.Member, err error) {
	srcDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	modulePath, err := getModulePath(srcDir)
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
		Dir:        srcDir,
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

	a.PkgFiles = make(map[*ssa.Package][]string)
	pkgFiles := make(map[*ssa.Package]map[string]struct{})
	fileMembers = make(map[string][]ssa.Member)

	for _, pkg := range pkgs {
		pkgFiles[pkg] = make(map[string]struct{})

		for _, member := range pkg.Members {
			switch member.Token() {
			// `pkg.Members` does not include methods,
			// use `graph.Nodes` to get all functions and methods
			case token.CONST, token.VAR, token.FUNC:
				continue
			}

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

			pkgFiles[pkg][filename] = struct{}{}
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
		if !strings.HasPrefix(filename, srcDir) {
			continue
		}

		if fn.Pkg != nil {
			pkgFiles[fn.Pkg][filename] = struct{}{}
		}
		fileMembers[filename] = append(fileMembers[filename], fn)
	}

	for pkg, files := range pkgFiles {
		for file := range files {
			a.PkgFiles[pkg] = append(a.PkgFiles[pkg], file)
		}
	}

	a.Prog = prog
	a.Callgraph = graph

	return
}

func getModulePath(dir string) (string, error) {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed to get module path: %v", err)
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
