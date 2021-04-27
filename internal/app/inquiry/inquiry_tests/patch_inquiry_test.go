package inquirytests

import (
	"context"
	"testing"

	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type PatchInquiryTestSuite struct {
	suite.Suite
	depCon container.Container
}

func (s *PatchInquiryTestSuite) SetupSuite() {
	manager.
		NewDefaultManager(context.Background()).
		Run(func() {
			deps.Get().Run()
			s.depCon = deps.Get().Container
		})
}

func (s *PatchInquiryTestSuite) PatchInquirySuccess() {
	//Create inquirer.

	// Create inquiry.
	//util.GenTestInquiryParams()

}

func TestPatchInquiryTestSuite(t *testing.T) {
	suite.Run(t, new(PatchInquiryTestSuite))
}
