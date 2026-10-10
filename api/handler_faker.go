package api

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type node struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Attributes  []string `json:"attributes,omitempty"`
	Weight      float64  `json:"weight,omitempty"`
	DependsOn   []string `json:"dependsOn,omitempty"`
	Children    []node   `json:"children,omitempty"`
	Custom      bool     `json:"custom,omitempty"`
	HasFakeFunc bool     `json:"hasFakeFunc,omitempty"`
}

func (h *handler) handleFakerTree(w http.ResponseWriter, r *http.Request) {
	n := generator.FindByName(generator.RootName)
	result := toNode(n, "")

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) setupFaker() {
	r := h.router.PathPrefix("/api/faker").Subrouter()

	r.HandleFunc("/tree", h.handleFakerTree).Methods(http.MethodGet)
	r.HandleFunc("/node/{name}/fake", h.handleFake).Methods(http.MethodGet, http.MethodPost)
}

func (h *handler) handleFake(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	n := generator.FindByName(vars["name"])
	if n == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var s *schema.Schema
	if r.Method == http.MethodPost {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		err = dynamic.UnmarshalJSON(b, &s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}

	req := generator.NewRequest(nil, s, nil)
	v, err := n.Fake(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
	}
	write(w, v)
}

func toNode(n *generator.Node, path string) node {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	if n.Name != generator.RootName {
		path += n.Name
	}

	var children []node
	for _, child := range n.Children {
		if child == nil {
			continue
		}
		children = append(children, toNode(child, path))
	}

	return node{
		Name:        n.Name,
		Path:        path,
		Attributes:  n.Attributes,
		Weight:      n.Weight,
		DependsOn:   n.DependsOn,
		Children:    children,
		Custom:      n.Custom,
		HasFakeFunc: n.Fake != nil,
	}
}
