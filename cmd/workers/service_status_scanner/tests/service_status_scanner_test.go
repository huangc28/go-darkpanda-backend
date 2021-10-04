package tests

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	cinternal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/stretchr/testify/suite"
)

type ServiceStatusScannerTestSuite struct {
	suite.Suite
	container cinternal.Container
}

func (s *ServiceStatusScannerTestSuite) SetupSuite() {
	ctx := context.Background()

	// Initialize app.
	manager.NewDefaultManager(ctx).Run(func() {
		if err := deps.Get().Run(); err != nil {
			log.Fatalf("failed to initialize dependency container %s", err.Error())
		}

		s.container = deps.Get().Container
	})
}

func (s *ServiceStatusScannerTestSuite) TestScanExpiredServices() {
	srvs := make([]models.Service, 0)

	for i := 0; i < 3; i++ {
		// Seed inquirer.
		maleParams, _ := util.GenTestUserParams()
		maleParams.Gender = models.GenderMale

		q := models.New(db.GetDB())

		ctx := context.Background()
		maleUser, err := q.CreateUser(ctx, *maleParams)

		if err != nil {
			s.Suite.T().Fatalf("failed to create inquirer %v", err)
		}

		// Seed picker.
		femaleParams, _ := util.GenTestUserParams()
		femaleParams.Gender = models.GenderFemale
		femaleUser, err := q.CreateUser(ctx, *femaleParams)

		if err != nil {
			s.Suite.T().Fatalf("failed to create picker %v", err)
		}

		iqParams, _ := util.GenTestInquiryParams(maleUser.ID)
		iqParams.PickerID = sql.NullInt32{
			Valid: true,
			Int32: int32(femaleUser.ID),
		}

		// Seed 3 inquiries.
		iq, err := q.CreateInquiry(ctx, *iqParams)

		if err != nil {
			s.Suite.T().Fatalf("failed to create inquiry %v", err)
		}

		// Seed 3 expired services with status `to_be_fulfilled`.
		serviceParams, err := util.GenTestServiceParams(maleUser.ID, femaleUser.ID, iq.ID)

		if err != nil {
			s.Suite.T().Fatalf("failed to create inquiry %v", err)
		}

		serviceParams.ServiceStatus = models.ServiceStatusToBeFulfilled

		now := time.Now()
		serviceParams.StartTime = sql.NullTime{
			Valid: true,
			Time:  now,
		}

		serviceParams.EndTime = sql.NullTime{
			Valid: true,
			Time:  now.AddDate(0, 0, -1),
		}

		srv, err := q.CreateService(ctx, *serviceParams)

		if err != nil {
			s.Suite.T().Fatalf("failed to create services %v", err)
		}

		srvs = append(srvs, srv)
	}

	var srvDao contracts.ServiceDAOer
	s.container.Make(&srvDao)

	esrvs, err := srvDao.ScanExpiredServices()

	if err != nil {
		s.Suite.T().Fatalf("failed to scan expired services %v", err)
	}

	srvUuidMap := make(map[string]bool)

	for _, srv := range srvs {
		srvUuidMap[srv.Uuid.String] = true
	}

	// assert that there are 3 expired services
	for _, esrv := range esrvs {
		_, ok := srvUuidMap[esrv.Uuid.String]

		s.Assert().True(ok, fmt.Sprintf("%s is expired", esrv.Uuid.String))
	}
}

func TestServiceStatusScannerSuite(t *testing.T) {
	suite.Run(t, new(ServiceStatusScannerTestSuite))
}
