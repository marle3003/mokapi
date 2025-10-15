package generator

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	nullFaker = &faker{fake: func() (any, error) {
		return nil, nil
	}}
)

type faker struct {
	fake fakeFunc
	node *Node
}

func newFakerWithFallback(n *Node, r *Request) *faker {
	return &faker{
		fake: func() (any, error) {
			if n.Fake == nil {
				log.Debugf("fake function for '%s' not defined", n.Name)
			}
			v, err := n.Fake(r)
			if err != nil {
				if errors.Is(err, NotSupported) {
					return fakeBySchema(r)
				}
				return nil, err
			}
			if v, err = validate(v, r); err != nil {
				return fakeBySchema(r)
			}
			return v, nil
		},
		node: n,
	}
}

func newFaker(f fakeFunc) *faker {
	return &faker{fake: f}
}

func fakeWithRetries(attempts int, fake func() error) error {
	attempt := 0
	var err error
	for {
		if attempt >= attempts {
			return fmt.Errorf("reached attempt limit (%d) caused by: %w", attempts, err)
		}

		err = fake()
		if err == nil {
			return nil
		}
		attempt++
	}
}
