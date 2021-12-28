package store

import (
	"mokapi/test"
	"testing"
)

func TestBroker(t *testing.T) {
	b := &Broker{id: 1, name: "foo", host: "bar", port: 9092}
	test.Equals(t, 1, b.Id())
	test.Equals(t, "foo", b.Name())
	test.Equals(t, "bar", b.Host())
	test.Equals(t, 9092, b.Port())
	test.Equals(t, "bar:9092", b.Addr())
}
