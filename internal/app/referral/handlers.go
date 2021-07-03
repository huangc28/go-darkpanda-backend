package referral

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	genverifycode "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/generate_verify_code"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
)

type VerifyReferalCodeBody struct {
	InviteeUUID  string `form:"invitee_uuid"`
	ReferralCode string `form:"referral_code"`
}

func HandleVerifyReferralCode(c *gin.Context, depCon container.Container) {
	var (
		body VerifyReferalCodeBody
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.FailedToValidateVerifyReferralCodeParams),
		)

		return
	}

	// Retrieve invitee ID
	var userSrv contracts.UserDAOer
	depCon.Make(&userSrv)

	invitee, err := userSrv.GetUserByUuid(body.InviteeUUID, "id")

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetUserIDByUuid,
				err.Error(),
			),
		)

		return
	}

	err, errCode := db.Transact(
		db.GetDB(),
		func(tx *sqlx.Tx) (error, interface{}) {
			refCodeDao := NewReferralCodeDAO(tx)

			refCode, err := refCodeDao.GetByRefCode(body.ReferralCode, []string{"*"})

			if err != nil {
				if err == sql.ErrNoRows {
					return err, apperr.ReferralCodeNotFound
				}

				return err, apperr.FailedToGetReferralCode
			}

			// If referral code is occupied returns error.
			if refCode.InviteeID.Valid {
				return errors.New("Referral code is occupied"), apperr.ReferralCodeIsOccupied
			}

			// If referral code is expired, return error.
			if time.Now().After(refCode.ExpiredAt.Time) {
				return errors.New(
					apperr.ReferralErrorMessageMap[apperr.ReferralCodeExpired],
				), apperr.ReferralCodeExpired
			}

			// Update the invitee of the referral code.
			err = refCodeDao.UpdateReferralCodeByID(UpdateReferralCodeParams{
				ID:        &refCode.ID,
				InviteeID: &invitee.ID,
				RefCode:   &body.ReferralCode,
			})

			if err != nil {
				return err, apperr.FailedToUpdateReferralcode
			}

			return nil, nil
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				errCode.(string),
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

func GetReferralCodeHandler(c *gin.Context, depCon container.Container) {
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)
	user, err := userDao.GetUserByUuid(c.GetString("uuid"), "id")

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

	rcDao := NewReferralCodeDAO(db.GetDB())
	refCode, err := rcDao.GetUnoccupiedReferralCode(c.GetString("uuid"))

	if err == sql.ErrNoRows {
		// Active referral code not found, we will create a fresh one here.
		refCode, err = rcDao.CreateReferralCode(
			CreateReferralCodeParams{
				InvitorID: int(user.ID),
				RefCode:   genverifycode.GenNum(100000, 999999),
			},
		)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToCreateReferralCode,
					err.Error(),
				),
			)

			return
		}
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetOccupiedRefcode,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct {
		ReferralCode string `json:"referral_code"`
	}{
		ReferralCode: refCode.RefCode,
	})
}
