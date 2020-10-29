package user

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	log "github.com/sirupsen/logrus"
)

type UserHandlers struct {
	PaymentDAO PaymentDAOer
	ServiceDAO ServiceDAOer
}

// Get the following information from the user:
//   - Gender
//   - Username
//   - Active inquiry
func (h *UserHandlers) GetMyProfileHandler(c *gin.Context) {
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
		data, err := gatherMaleInfo(
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
		// gather the following female information.
		//   - username
		//   - gender
		//   - uuid
		//   - avatar url

		log.Printf("DEBUG ctrl 1 %v", usr.AvatarUrl.String)
		c.JSON(http.StatusOK, NewTransform().TransformUser(&usr))
	}

}

func gatherMaleInfo(ctx context.Context, q *models.Queries, usr *models.User) (*TransformUserWithInquiryData, error) {
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

type GetUserProfileBody struct {
	UUID string `form:"uuid" json:"uuid" binding:"required,gt=0"`
}

func (h *UserHandlers) GetUserProfileHandler(c *gin.Context) {
	var (
		uuid string          = c.Param("uuid")
		ctx  context.Context = context.Background()
	)

	q := models.New(db.GetDB())
	user, err := q.GetUserByUuid(ctx, uuid)

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

	tResp, err := NewTransform().TransformMaleUser(user)

	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, tResp)
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

func (h *UserHandlers) PutUserInfo(c *gin.Context) {
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
	user, err := dao.UpdateUserInfoByUuid(ctx, contracts.UpdateUserInfoParams{
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

func (h *UserHandlers) PatchUserImages(c *gin.Context) {
	c.JSON(http.StatusOK, struct{}{})
}

type GetUserImagesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=9"`
}

func (h *UserHandlers) GetUserImagesHandler(c *gin.Context) {
	uuid := c.Param("uuid")

	body := &GetUserImagesBody{}

	if err := requestbinder.Bind(c, body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateGetUserImagesParams,
				err.Error(),
			),
		)

		return
	}

	// Get image link by user uuid
	images, err := NewUserDAO(db.GetDB()).GetUserImagesByUuid(
		uuid,
		body.Offset,
		body.PerPage,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetImagesByUserUUID,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformUserImages(images))
}

func (h *UserHandlers) GetUserPayments(c *gin.Context) {
	uuid := c.Param("uuid")

	paymentInfos, err := h.PaymentDAO.GetPaymentsByUuid(uuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserPayments,
				err.Error(),
			),
		)

		return

	}

	trfmPaymentInfo, err := NewTransform().TransformPaymentInfo(paymentInfos)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformUserPayments,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfmPaymentInfo)
}

type GetUserServiceHistoryRecords struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=5"`
}

func (h *UserHandlers) GetUserServiceHistory(c *gin.Context) {
	uuid := c.Param("uuid")
	body := GetUserServiceHistoryRecords{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateGetServiceHistoryParams,
				err.Error(),
			),
		)

		return
	}

	// Retrieve past service records.
	services, err := h.ServiceDAO.GetUserHistoricalServicesByUuid(
		uuid,
		body.PerPage,
		body.Offset,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetHistoricalServices,
				err.Error(),
			),
		)

		return
	}

	trfmSrvs, err := NewTransform().TransformHistoricalServices(services)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformHistoricalServices,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfmSrvs)
}
