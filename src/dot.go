package main

import (
	"bytes"
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
}`

const tmplTable = `{{define "table"}}
{{.Id}} [id="{{.Id}}", label=<
	<TABLE BORDER="0" CELLBORDER="0">
	<TR><TD WIDTH="230" BORDER="0"><FONT POINT-SIZE="12">{{.Title}}</FONT></TD></TR>

	{{range .Sections}}
	<TR><TD>
	<TABLE BORDER="0" CELLSPACING="0" CELLPADDING="4" CELLBORDER="1">
	{{template "cell" .}}
	</TABLE>
	</TD></TR>
	{{end}}

	<TR><TD BORDER="0"></TD></TR>
	</TABLE>
>];
{{end}}`

const tmplCell = `{{define "cell"}}
<TR><TD PORT="{{.Id}}" ID="{{.Id}}">{{.Title}}</TD></TR>
{{range .SubNodes}}
{{template "cell" .}}
{{end}}
{{end}}`

const tmplEdge = `{{define "edge"}}
{{.From.TableID}}:{{.From.NodeID}} -> {{.To.TableID}}:{{.To.NodeID}}
{{end}}`

const tmplCluster = `{{define "cluster"}}
subgraph cluster_{{.Title}} {
	label = "{{.Title}}";

	{{.Nodes}}

	{{.Subgraph}}
};
{{end}}`

func renderDot(g *Graph) (dot string, err error) {
	t := template.New("dot")
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
