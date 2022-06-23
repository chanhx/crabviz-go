package main

import (
	"crabviz-go/app"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
)

//go:embed app/static
var static embed.FS

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		log.Fatal("need a path argument")
	}

	app := app.NewApp(args[0])

	addr := ":8090"
	log.Printf("http serving at %s", addr)

	fSys, err := fs.Sub(static, "app/static")
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fSys))))
	http.Handle("/", app)
	http.ListenAndServe(addr, nil)
}
