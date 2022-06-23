package graph

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"text/template"
)

const tmplGraph = `digraph {
	graph [
		rankdir = "LR"
		ranksep = 2.0
	];
	node [
		fontsize = "16"
		fontname = "helvetica, open-sans"
		shape = "plaintext"
		style = "rounded, filled"
	];

	{{range .Tables}}
	{{template "table" .}}
	{{end}}

	{{range .Edges}}
	{{template "edge" .}}
	{{end}}

	{{range .Clusters}}
	{{template "cluster" .}}
	{{end}}
}`

const tmplTable = `{{define "table"}}
{{.Id}} [id="{{.Id}}", label=<
	<TABLE BORDER="0" CELLBORDER="0">
	<TR><TD WIDTH="230" BORDER="0"><FONT POINT-SIZE="12">{{.Title}}</FONT></TD></TR>

	{{range .Sections}}
	<TR><TD>
	<TABLE BORDER="0" CELLSPACING="0" CELLPADDING="4" CELLBORDER="1">
	{{template "cell" cell . $.Id}}
	</TABLE>
	</TD></TR>
	{{end}}

	<TR><TD BORDER="0"></TD></TR>
	</TABLE>
>];
{{end}}`

const tmplCell = `{{define "cell"}}
<TR><TD PORT="{{.Id}}" ID="{{.TableID}}:{{.Id}}" HREF="remove_me_url.cell{{.ClassesStr}}">{{.FormattedTitle}}</TD></TR>
{{range .SubNodes}}
{{template "cell" cell . $.TableID}}
{{end}}
{{end}}`

const tmplEdge = `{{define "edge"}}
{{.From.TableID}}:{{.From.NodeID}} -> {{.To.TableID}}:{{.To.NodeID}} [{{.Attributes}}]
{{end}}`

const tmplCluster = `{{define "cluster"}}
subgraph "cluster_{{.Title}}" {
	label = "{{.Title}}";

	{{range .Nodes}}
	{{.}}
	{{end}}

	{{range .SubClusters}}
	{{template "cluster" .}}
	{{end}}
};
{{end}}`

func RenderDot(g *Graph) (dot string, err error) {
	t := template.New("dot").Funcs(template.FuncMap{"cell": cell})

	for _, s := range []string{tmplGraph, tmplTable, tmplCell, tmplEdge, tmplCluster} {
		if _, err = t.Parse(s); err != nil {
			return
		}
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, g); err != nil {
		return
	}

	dot = buf.String()

	return
}

type Cell struct {
	Node
	TableID uint32
}

func cell(node Node, tableID uint32) Cell {
	return Cell{node, tableID}
}

func (node Node) FormattedTitle() string {
	title := node.Title

	if strings.LastIndex(title, "]") == -1 {
		return title
	}

	start := strings.IndexRune(title, '[')

	args := strings.Split(title[start:], " ")

	return fmt.Sprintf("%s<BR/>%s", title[:start], strings.Join(args, "<BR/>"))
}

func (node Node) ClassesStr() string {
	if len(node.Classes) == 0 {
		return ""
	}
	return "." + strings.Join(node.Classes, ".")
}

func (e Edge) Attributes() string {
	var attrs []string

	id := fmt.Sprintf(`id="%d:%d -> %d:%d"`, e.From.TableID, e.From.NodeID, e.To.TableID, e.To.NodeID)
	attrs = append(attrs, id)

	if e.From.TableID == e.To.TableID {
		attrs = append(attrs, `class="modify-me"`)
	}

	switch e.Style {
	case Dashed:
		attrs = append(attrs, `style="dashed"`)
	case DotArrow:
		attrs = append(attrs, `arrowhead="normalnoneodot"`)
	}

	return strings.Join(attrs, ", ")
}

func DotExport(dot string, format string) ([]byte, error) {
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = bytes.NewReader([]byte(dot))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("failed to export %s: %v", format, err)
	}

	return output, nil
}

func DotExportSVG(dot string) (string, error) {
	output, err := DotExport(dot, "svg")
	if err != nil {
		return "", err
	}

	return string(output), nil
}
