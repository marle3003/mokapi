package generator

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type faker struct {
	fake fakeFunc
	node *Node
	// tokens are the path tokens node was matched with
	tokens []string
}

func newFakerWithFallback(n *Node, tokens []string, r *Request) *faker {
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
				if n.Custom {
					log.Debugf("custom fake function at node failed for '%s': %v", n.Name, err)
				}
				return fakeBySchema(r)
			}
			if n != g.root.defaultNode {
				r.recordNode(n, tokens, v)
			}
			return v, nil
		},
		node:   n,
		tokens: tokens,
	}
}

func newFaker(f fakeFunc) *faker {
	return &faker{fake: f}
}

func newNullFaker(r *Request) *faker {
	return newFaker(func() (any, error) {
		r.recordSource(SourceNull, nil)
		return nil, nil
	})
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
