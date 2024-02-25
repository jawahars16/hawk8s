package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jawahars16/hawk8s/internal/core"
	"github.com/jawahars16/hawk8s/internal/kubeclient"
)

func main() {
	port := flag.String("port", "3000", "Port to run the server")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	r := chi.NewRouter()
	if *verbose {
		r.Use(middleware.Logger)
	}

	kubeclient := kubeclient.NewKubeClient()

	tmpl := template.Must(template.New("").Funcs(sprig.FuncMap()).ParseGlob("internal/templates/*.html"))
	coreService := core.NewService(kubeclient)
	coreHandler := core.NewHandler(tmpl, coreService)

	r.Get("/", coreHandler.GetIndex)
	r.Get("/nodes", coreHandler.GetNodes)
	r.Get("/pods", coreHandler.GetPods)
	r.Get("/namespaces", coreHandler.GetNamespaces)

	fileServer(r)
	log.Println("Starting server at :3000")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), r); err != nil {
		log.Fatal(err)
	}
}

func fileServer(r chi.Router) {
	fs := http.StripPrefix("/static", http.FileServer(http.Dir("./static")))
	r.Get("/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		fs.ServeHTTP(w, r)
	}))
}
