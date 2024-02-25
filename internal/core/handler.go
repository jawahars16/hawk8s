package core

import (
	"context"
	"html/template"
	"net/http"
)

type service interface {
	// GetViewModel(ctx context.Context, namespace string, mode string) *viewModel
	GetNamespaces(ctx context.Context) ([]namespace, error)
	GetNodes(ctx context.Context) ([]node, error)
	GetPods(ctx context.Context, node string) ([]pod, error)
}

type Handler struct {
	tmpl            *template.Template
	service         service
	activeMode      string
	activeNamespace string
}

func NewHandler(tmpl *template.Template, service service) *Handler {
	return &Handler{
		tmpl:            tmpl,
		service:         service,
		activeMode:      CPU,
		activeNamespace: "all",
	}
}

func (h *Handler) GetIndex(w http.ResponseWriter, r *http.Request) {
	err := h.tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces, err := h.service.GetNamespaces(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.tmpl.ExecuteTemplate(w, "filter.html", namespaces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.service.GetNodes(r.Context())
	vm := nodeViewModel{
		Nodes: nodes,
		Title: "Nodes",
	}
	if err != nil {
		vm.Error = err.Error()
	}
	err = h.tmpl.ExecuteTemplate(w, "core.html", vm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetPods(w http.ResponseWriter, r *http.Request) {
	node := r.URL.Query().Get("node")
	pods, err := h.service.GetPods(r.Context(), node)
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}
	err = h.tmpl.ExecuteTemplate(w, "pods.html", podViewModel{
		Pods:  pods,
		Error: errorMsg,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
