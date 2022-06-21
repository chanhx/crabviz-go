package main

import (
	"crabviz-go/src/app"
	"log"
	"net/http"
)

func main() {
	app := app.NewApp()

	addr := ":8090"
	log.Printf("http serving at %s", addr)

	http.ListenAndServe(addr, app)
}
