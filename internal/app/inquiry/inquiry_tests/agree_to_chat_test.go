package inquirytests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AgreeToChatTestSuite struct {
	suite.Suite
}

func (s *AgreeToChatTestSuite) TestAgreeToChatSuccess() {
	//  create
	//
}

func TestAgreeToChatTestSuite(t *testing.T) {
	suite.Run(t, new(AgreeToChatTestSuite))
}
