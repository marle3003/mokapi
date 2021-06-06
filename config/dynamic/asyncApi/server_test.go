package asyncApi

import (
	"testing"
)

func TestGetHost(t *testing.T) {
	server := Server{Url: "demo:9092"}
	if s := server.GetHost(); s != "demo" {
		t.Errorf("Parse(%q): got %v, expected demo", "demo:9092", s)
	}
}

func TestGetPort(t *testing.T) {
	server := Server{Url: "demo:9092"}
	if p := server.GetPort(); p != 9092 {
		t.Errorf("Parse(%q): got %v, expected 9092", "demo:9092", p)
	}
}
