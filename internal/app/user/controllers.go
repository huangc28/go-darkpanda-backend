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

type PutUserInfoBody struct {
	AvatarURL   *string  `json:"avatar_url"`
	Nationality *string  `json:"nationality"`
	Region      *string  `json:"region"`
	Age         *int     `json:"age"`
	Height      *float64 `json:"height"`
	Weight      *float64 `json:"weight"`
	Habbits     *string  `json:"habbits"`
	Description *string  `json:"description"`
	BreastSize  *string  `json:"breast_size"`
}

func PutUserInfo(c *gin.Context) {
	body := &PutUserInfoBody{}

	if err := c.ShouldBindJSON(body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidatePutUserParams,
				err.Error(),
			),
		)

		return
	}

	// ------------------- Update user info -------------------
	ctx := context.Background()
	uuid := c.GetString("uuid")
	dao := NewUserDAO(db.GetDB())
	user, err := dao.UpdateUserInfoByUuid(ctx, UpdateUserInfoParams{
		AvatarURL:   body.AvatarURL,
		Nationality: body.Nationality,
		Region:      body.Region,
		Age:         body.Age,
		Height:      body.Height,
		Weight:      body.Weight,
		Description: body.Description,
		BreastSize:  body.BreastSize,
		Uuid:        uuid,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"uuid": uuid,
		}).Errorf("Failed to patch user info by uuid %s", err.Error())

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToPatchUserInfo,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformPatchedUser(user))
}

func PatchUserImages(c *gin.Context) {
	c.JSON(http.StatusOK, struct{}{})
}
