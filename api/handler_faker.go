package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
	"net/http"
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
	result := toNode(n)

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func (h *handler) setupFaker() {
	r := h.router.PathPrefix("/api/faker").Subrouter()

	r.HandleFunc("/tree", h.handleFakerTree).Methods(http.MethodGet)
	r.HandleFunc("/fake", h.handleFakerExplain).Methods(http.MethodPost)
	r.HandleFunc("/node/fake", h.handleFakeNode).Methods(http.MethodGet, http.MethodPost)
}

type explainRequest struct {
	Path   []string        `json:"path"`
	Schema json.RawMessage `json:"schema"`
}

func (h *handler) handleFakerExplain(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	var body explainRequest
	if len(b) > 0 {
		if err = json.Unmarshal(b, &body); err != nil {
			writeBadRequest(w, err)
			return
		}
	}
	s, err := parseSchema(body.Schema)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	result, err := generator.Explain(generator.NewRequest(body.Path, s, nil))
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	write(w, result)
}

func (h *handler) handleFakeNode(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	n := generator.FindByPath(path)
	if n == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if n.Fake == nil {
		writeBadRequest(w, fmt.Errorf("node %s has no fake function", path))
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	s, err := parseSchema(b)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	v, err := n.Fake(generator.NewRequest(nil, s, nil))
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	write(w, v)
}

// parseSchema returns no schema for an empty body.
func parseSchema(b []byte) (*schema.Schema, error) {
	if len(b) == 0 {
		return nil, nil
	}
	var s *schema.Schema
	if err := dynamic.UnmarshalJSON(b, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func writeBadRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(err.Error()))
}

func toNode(n *generator.Node) node {
	var children []node
	for _, child := range n.Children {
		if child == nil {
			continue
		}
		children = append(children, toNode(child))
	}

	return node{
		Name:        n.Name,
		Path:        generator.NodePath(n),
		Attributes:  n.Attributes,
		Weight:      n.Weight,
		DependsOn:   n.DependsOn,
		Children:    children,
		Custom:      n.Custom,
		HasFakeFunc: n.Fake != nil,
	}
}
