package main

import (
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jawahars16/kubemon/internal/cache"
	"github.com/jawahars16/kubemon/internal/core"
	"github.com/jawahars16/kubemon/internal/kubeclient"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	cache := cache.New(5)
	kubeclient := kubeclient.NewKubeClient(cache)
	// mockKubeClient := kubeclient.NewMockKube()
	// nodesService := nodes.NewService(kubeclient)

	tmpl := template.Must(template.New("").Funcs(sprig.FuncMap()).ParseGlob("../internal/templates/*.html"))
	// indexHandler := index.NewIndexHandler(tmpl)
	// filterHandler := filter.NewHandler(tmpl, kubeclient)
	// nodesHandler := nodes.NewNodesHandler(tmpl, nodesService)
	coreService := core.NewService(kubeclient)
	coreHandler := core.NewHandler(tmpl, coreService)

	r.Get("/", coreHandler.GetIndex)
	r.Get("/nodes", coreHandler.GetNodes)
	r.Get("/pods", coreHandler.GetPods)
	r.Get("/namespaces", coreHandler.GetNamespaces)

	// r.Get("/nodes", nodesHandler.Get)
	// r.Get("/pod/info", podHandler.Get)

	fileServer(r)
	http.ListenAndServe(":3000", r)
}

func fileServer(r chi.Router) {
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("../static")))
	r.Get("/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
