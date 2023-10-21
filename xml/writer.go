package xml

import (
	"encoding/xml"
	"fmt"
	"io"
)

func (n *Node) Write(w io.Writer) {
	n.WriteStart(w)
	n.WriteContent(w)
	n.WriteEnd(w)
}

func (n *Node) WriteStart(w io.Writer) {
	writeString(w, "<"+n.Name)
	n.writeAttributes(w)
	writeString(w, ">")
}

func (n *Node) WriteContent(w io.Writer) {
	_ = xml.EscapeText(w, []byte(n.Content))
}

func (n *Node) WriteEnd(w io.Writer) {
	writeString(w, fmt.Sprintf("</%v>", n.Name))
}

func (n *Node) writeAttributes(w io.Writer) {
	for it := n.Attributes.Iter(); it.Next(); {
		writeString(w, fmt.Sprintf(" %v=\"%v\"", it.Key(), it.Value()))
	}
}

func writeString(w io.Writer, s string) {
	_, _ = w.Write([]byte(s))
}
