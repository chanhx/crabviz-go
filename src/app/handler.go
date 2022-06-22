package app

import (
	"log"
	"net/http"
	"text/template"

	"crabviz-go/src/analysis"
	"crabviz-go/src/graph"
)

const tmpl = `
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" type="text/css" href="static/styles.css">
    <script src="static/path-data-polyfill.min.js"></script>
    <script src="static/svg-pan-zoom.min.js"></script>
</head>
<body>
    {{.}}
    <script src="static/preprocess.js"></script>
</body>
</html>
`

type App struct {
	analyzer *analysis.Analyzer
}

func NewApp() *App {
	return &App{
		analyzer: new(analysis.Analyzer),
	}
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileMembers, err := app.analyzer.Analyze("", false, []string{})
	if err != nil {
		panic(err)
	}

	g := graph.GenGraph(app.analyzer.Prog.Fset, fileMembers, app.analyzer.Callgraph)

	dot, err := graph.RenderDot(&g)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	svg, err := graph.DotExportSVG(dot)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	t := template.New("graph")
	if _, err = t.Parse(tmpl); err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err = t.ExecuteTemplate(w, "graph", svg); err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
}
