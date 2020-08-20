package user

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	log "github.com/sirupsen/logrus"
)

// Get the following information from the user:
//   - Gender
//   - Username
//   - Active inquiry
func GetUserInfo(c *gin.Context) {
	// ------------------- retrieve user model -------------------
	var (
		uuid string          = c.GetString("uuid")
		ctx  context.Context = context.Background()
	)

	tx, err := db.
		GetDB().
		BeginTx(ctx, nil)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBeginTx,
				err.Error(),
			),
		)

		tx.Rollback()
		return
	}

	q := models.New(tx)
	usr, err := q.GetUserByUuid(ctx, uuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				err.Error(),
			),
		)

		return
	}

	// ------------------- get user relative info base on gender -------------------
	switch usr.Gender {
	case models.GenderMale:
		data, err := GatherMaleInfo(
			ctx,
			q,
			&usr,
		)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToFindInquiryByInquiererID,
					err.Error(),
				),
			)

			tx.Rollback()
			return

		}

		if err := tx.Commit(); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToFindInquiryByInquiererID,
					err.Error(),
				),
			)

			return
		}

		c.JSON(http.StatusOK, data)

	case models.GenderFemale:
		log.Printf("gather female info")
	}

}

func GatherMaleInfo(ctx context.Context, q *models.Queries, usr *models.User) (*TransformUserWithInquiryData, error) {
	// ------------------- check if user has an active service -------------------
	inquiry, err := q.GetInquiryByInquirerID(ctx, models.GetInquiryByInquirerIDParams{
		InquirerID: sql.NullInt32{
			Int32: int32(usr.ID),
			Valid: true,
		},
		InquiryStatus: models.InquiryStatusInquiring,
	})

	if err != nil {
		if err != sql.ErrNoRows {

			return nil, err
		}

	}

	inquiries := make([]*models.ServiceInquiry, 0)

	if err != sql.ErrNoRows {
		inquiries = append(inquiries, &inquiry)
	}

	return NewTransform().TransformUserWithInquiry(usr, inquiries), nil
}
