package acceptance

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestAcceptance(t *testing.T) {
	suite.Run(t, new(PetStoreSuite))
}
