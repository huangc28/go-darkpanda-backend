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
