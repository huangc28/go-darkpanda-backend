package util

import (
	"context"
	"database/sql"
	"fmt"

	faker "github.com/bxcodec/faker/v3"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/ventu-io/go-shortid"
)

func randomGender() models.Gender {
	gs := []models.Gender{
		models.GenderFemale,
		models.GenderMale,
	}

	return gs[seededRand.Intn(len(gs))]
}

// GenTestUser generate randomized data on user fields but now create it.
func GenTestUserParams(ctx context.Context) (*models.CreateUserParams, error) {
	p := &models.CreateUserParams{}
	if err := faker.FakeData(p); err != nil {
		return nil, err
	}

	sid, err := shortid.Generate()

	if err != nil {
		return nil, err
	}

	p.Username = faker.Username()
	p.Uuid = sid
	p.Gender = randomGender()
	p.PremiumType = models.PremiumTypeNormal
	p.PhoneVerifyCode = sql.NullString{
		String: fmt.Sprintf("%s-%d", GenRandStringRune(3), Gen4DigitNum(1000, 9999)),
		Valid:  true,
	}

	return p, nil
}
