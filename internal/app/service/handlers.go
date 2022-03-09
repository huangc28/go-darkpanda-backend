package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	dpfcm "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/firebase_messaging"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// GetListOfCurrentServicesHandler retrieve those service of the following status:
//   - unpaid
//   - to_be_fulfilled
//
// Note 2022/03/08: we are removing DP point deposit here. We are not retrieving unpaid services
type GetListOfCurrentServicesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=5"`
}

func GetIncomingServicesHandler(c *gin.Context, depCon container.Container) {
	body := GetListOfCurrentServicesBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindInquiryUriParams,
				err.Error(),
			),
		)

		return
	}

	// Retrieve picker's uuid
	userUuid := c.GetString("uuid")

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(
		userUuid,
		"id",
		"gender",
	)

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

	// Retrieve list of incoming services
	var srvs []ServiceResult = make([]ServiceResult, 0)

	srvDao := NewServiceDAO(db.GetDB())

	// Note 2022/03/08: Since we remove DP point deposit, we will only retrieve service
	// with status `to_be_fulfilled`
	srvs, err = srvDao.GetServicesByStatus(
		int(user.ID),
		user.Gender,
		body.Offset,
		body.PerPage,
		models.ServiceStatusToBeFulfilled,
		models.ServiceStatusFulfilling,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetIncomingService,
				err.Error(),
			),
		)

		return
	}

	// Retrieve latest message for each chatroom. Collect slice of chatroom uuids.
	channelUuids := make([]string, 0)
	ctx := context.Background()

	for _, srv := range srvs {
		channelUuids = append(channelUuids, srv.ChannelUuid.String)
	}

	df := darkfirestore.Get()
	msgs, err := df.GetLatestMessageForEachChatroom(ctx, channelUuids)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMessageFromFireStore,
				err.Error(),
			),
		)

		return
	}

	// Retrieve service provider uuid
	c.JSON(
		http.StatusOK,
		TransformGetIncomingServices(
			srvs,
			msgs,
		),
	)
}

type GetOverduedServicesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"per_page,default=5"`
}

// GetOverduedServicesHandlers retrieve those service of the following status:
//  - canceled
//  - failed_due_to_both
//  - failed_due_to_girl
//  - failed_due_to_man
//  - completed
func GetOverduedServicesHandlers(c *gin.Context, depCon container.Container) {
	body := GetOverduedServicesBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return
	}

	// Retrieve picker's uuid
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(
		c.GetString("uuid"),
		"id",
		"gender",
	)

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

	var srvRes []ServiceResult = make([]ServiceResult, 0)

	// Retrieve list of overdued services
	srvDao := NewServiceDAO(db.GetDB())
	srvRes, err = srvDao.GetServicesByStatus(
		int(user.ID),
		user.Gender,
		body.Offset,
		body.PerPage,
		models.ServiceStatusCanceled,
		models.ServiceStatusCompleted,
		models.ServiceStatusExpired,
		models.ServiceStatusPaymentFailed,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetOverdueService,
				err.Error(),
			),
		)

		return
	}

	c.JSON(
		http.StatusOK,
		TransformOverDueServices(srvRes),
	)
}

type ScanServiceQrCodeBody struct {
	QrcodeSecret string `json:"qrcode_secret" form:"qrcode_secret" binding:"required,gt=0"`
	QrcodeUuid   string `json:"qrcode_uuid" form:"qrcode_uuid" binding:"required,gt=0"`
}

// ScanServiceQrCode Upon meetup, female / male user would scan service QR code to
// start fulfilling the service. This API checks the following conditions before
// change the service status to `fulfilling`:
//   1. Check if the current datetime is within the range of appointment time.
//   2. If the scanner is one of the participants of this service. The scanner has
//      to be either inquirer or service provider.
//   3. the service status is `to_be_fulfilled`
func ScanServiceQrCode(c *gin.Context, depCon container.Container) {
	body := ScanServiceQrCodeBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return
	}

	if body.QrcodeSecret != config.GetAppConf().ServiceQrCodeSecret {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedServiceQrCodeSecretNotMatch,
			),
		)

		return

	}

	// Retrieve service by qrcode uuid.
	srvDao := NewServiceDAO(db.GetDB())
	srv, err := srvDao.GetServiceByQrcodeUuid(body.QrcodeUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByQrCodeUuid,
				err.Error(),
			),
		)

		return
	}

	// Check if current time is in between service appointment and the buffer time.
	err = IsTimeInRange(
		srv.AppointmentTime.Time,
		srv.AppointmentTime.Time.Add(30*time.Minute),
	)

	df := darkfirestore.Get()
	ctx := context.Background()

	srvStatusExp := models.ServiceStatusExpired
	if errors.Is(err, ErrorExpired) {
		// Update service status in firestore.
		df.UpdateService(
			ctx,
			darkfirestore.UpdateServiceParams{
				ServiceUuid:   srv.Uuid.String,
				ServiceStatus: srv.ServiceStatus.ToString(),
			},
		)

		// Change service type to expired.
		srvDao.UpdateServiceByID(contracts.UpdateServiceByIDParams{
			ID:            srv.ID,
			ServiceStatus: &srvStatusExp,
		})
	}

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.ServiceStartTimeNotValid,
				err.Error(),
			),
		)

		return
	}

	// Check if qrcode scanner is either inquirer or service provider.
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)
	user, err := userDao.GetUserByUuid(
		c.GetString("uuid"),
		"id",
	)

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

	if srv.ServiceProviderID.Int32 != int32(user.ID) && srv.CustomerID.Int32 != int32(user.ID) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.NotAServiceParticipant),
		)

		return
	}

	// Makesure the service status is `to_be_fulfilled`
	if srv.ServiceStatus != models.ServiceStatusToBeFulfilled {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.InvalidServiceStatus,
				fmt.Sprintf(
					"Invalid service status: %s. Unable to change to 'to_be_fulfilled' status",
					srv.ServiceStatus.ToString(),
				),
			),
		)

		return
	}

	// Change service status to `to_be_fulfilled`
	fsm := NewServiceFSM(srv.ServiceStatus)
	if err := fsm.Event(StartService.ToString()); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToChangeServiceStatus,
				err.Error(),
			),
		)

		return
	}

	srvStatus := models.ServiceStatus(fsm.Current())
	startTime := time.Now().UTC()
	endTime := startTime.Add(
		time.Duration(srv.Duration.Int32) * time.Minute,
	)

	var chatDao contracts.ChatDaoer
	depCon.Make(&chatDao)
	chat, err := chatDao.GetChatroomByServiceId(int(srv.ID))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChatroomByServiceId,
				err.Error(),
			),
		)
		return
	}

	// Change service status to be `fulfilling`.
	transResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			srvDao.WithTx(tx)
			usrv, err := srvDao.UpdateServiceByID(contracts.UpdateServiceByIDParams{
				ID:            srv.ID,
				ServiceStatus: &srvStatus,
				StartTime:     &startTime,
				EndTime:       &endTime,
			})

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToUpdateService,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			ctx := context.Background()
			err = df.StartService(ctx, darkfirestore.StartServiceParams{
				ServiceUuid:   usrv.Uuid.String,
				ChannelUuid:   chat.ChannelUuid.String,
				ServiceStatus: usrv.ServiceStatus,
				Data: darkfirestore.ChatMessage{
					From:    c.GetString("uuid"),
					Content: "",
				},
			})

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToStartService,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			return db.FormatResp{
				Response: usrv,
			}
		},
	)

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.ErrCode,
				transResp.Err.Error(),
			),
		)
		return
	}

	usrv := transResp.Response.(*models.Service)

	c.JSON(http.StatusOK, TransformScanServiceQrCode(usrv))
}

type GetServiceQRCodeBody struct {
	ServiceUuid string `uri:"seg"`
}

func GetServiceQRCode(c *gin.Context, depCon container.Container) {
	body := GetServiceQRCodeBody{}

	if err := c.ShouldBindUri(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return

	}

	// Retrieve QRCode url by service uuid.
	srvDao := NewServiceDAO(db.GetDB())
	qrCode, err := srvDao.GetQrcodeByServiceUuid(body.ServiceUuid)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.NoQRCodeFound,
					err.Error(),
				),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetQrCodeByServiceUuid,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct {
		QrCodeUrl string `json:"qrcode_url"`
	}{qrCode.Url.String})
}

func GetAvailableServices(c *gin.Context) {
	srvDao := NewServiceDAO(db.GetDB())

	srvOptions, err := srvDao.GetServiceOptions()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceNames,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, TransformServiceOptions(srvOptions))
}

func GetServicePaymentDetails(c *gin.Context, depCon container.Container) {
	srvUuid := c.Param("seg")

	// Retrieve payment detail of the service.
	var srvDao contracts.ServiceDAOer
	depCon.Make(&srvDao)

	srv, err := srvDao.GetServiceByUuid(srvUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return
	}

	var (
		pDaoer       contracts.PaymentDAOer
		rateDaoer    contracts.RateDAOer
		userDaoer    contracts.UserDAOer
		coinPkgDaoer contracts.CoinPackageDAOer
	)

	depCon.Make(&pDaoer)
	depCon.Make(&rateDaoer)
	depCon.Make(&userDaoer)
	depCon.Make(&coinPkgDaoer)

	user, err := userDaoer.GetUserByUuid(c.GetString("uuid"))

	if err != nil {
		c.AbortWithError(
			http.StatusNotFound,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				err.Error(),
			),
		)

		return
	}

	// Check if service has been commented before.
	hasCommented, err := rateDaoer.HasCommented(
		int(srv.ID),
		int(user.ID),
	)

	if err == sql.ErrNoRows {
		c.AbortWithError(
			http.StatusNotFound,
			apperr.NewErr(
				apperr.AssetNotFound,
				err.Error(),
			),
		)

		return
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckHasCommented,
				err.Error(),
			),
		)

		return
	}

	// Check against service partner id to see if the requester has blocked this guy.
	var blockDao contracts.BlockDAOer
	depCon.Make(&blockDao)
	srvPartnerId := srv.GetPartnerId(user.ID)

	hasBlocked, err := blockDao.HasBlockedByUserById(
		contracts.HasBlockedByUserByIdParams{
			BlockerId: int(user.ID),
			BlockeeId: int(srvPartnerId),
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckHasBlockedUser,
				err.Error(),
			),
		)

		return
	}

	p, err := pDaoer.GetPaymentByServiceUuid(srvUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetPaymentByServiceUuid,
				err.Error(),
			),
		)

		return
	}

	matchingFeeRate, err := coinPkgDaoer.GetMatchingFeeRate()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMatchingFee,
				err.Error(),
			),
		)
		return
	}

	matchingFee, err := matchingFeeRate.CalcMatchingFeeFromString(srv.Price.String)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCalcServiceMatchingFee,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, TrfPaymentDetail(
		TrfPaymentDetailParams{
			PaymentDetail: p,
			HasCommented:  hasCommented,
			HasBlocked:    hasBlocked,
			MatchingFee:   matchingFee,
		},
	))
}

func GetServiceDetailHandler(c *gin.Context, depCon container.Container) {
	serviceUuid := c.Param("seg")

	var (
		srvDao  contracts.ServiceDAOer
		coinDao contracts.CoinPackageDAOer
	)
	depCon.Make(&srvDao)
	depCon.Make(&coinDao)

	srv, err := srvDao.GetServiceByUuid(serviceUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return
	}

	matchingFeeRate, err := coinDao.GetMatchingFeeRate()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMatchingFee,
				err.Error(),
			),
		)

		return
	}

	matchingFee, err := matchingFeeRate.CalcMatchingFeeFromString(srv.Price.String)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCalcServiceMatchingFee,
				err.Error(),
			),
		)

		return

	}

	trfed, err := TrfServiceDetail(*srv, matchingFee)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformResponse,
				err.Error(),
			),
		)

		return

	}

	c.JSON(http.StatusOK, trfed)
}

func GetServiceRating(c *gin.Context, depCon container.Container) {
	var (
		srvUuid  string = c.Param("seg")
		userUuid string = c.GetString("uuid")
		userDao  contracts.UserDAOer
	)

	depCon.Make(&userDao)
	user, err := userDao.GetUserByUuid(userUuid)

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

	var rateDao contracts.RateDAOer
	depCon.Make(&rateDao)

	// Get service rating made by the chat partner.
	srvRating, err := rateDao.GetServiceRating(
		contracts.GetServiceRatingParams{
			ServiceUuid: srvUuid,
			RaterId:     int(user.ID),
		},
	)

	// if errors

	if err != nil {
		if err != sql.ErrNoRows {
			c.AbortWithError(
				http.StatusNotFound,
				apperr.NewErr(apperr.AssetNotFound),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceRating,
				err.Error(),
			),
		)

		return
	}

	c.JSON(
		http.StatusOK,
		TransformRate(srvRating),
	)
}

type CreateServiceRatingparams struct {
	ServiceUuid string `json:"service_uuid" form:"service_uuid" binding:"required,gt=0"`
	Rating      int    `json:"rating" form:"rating" binding:"required,gt=0"`
	Comment     string `json:"comment" form:"comment"`
}

func CreateServiceRating(c *gin.Context, depCon container.Container) {
	var (
		body    CreateServiceRatingparams
		srvUuid string = c.Param("seg")
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return
	}

	var userDao contracts.UserDAOer

	depCon.Make(&userDao)

	usr, err := userDao.GetUserByUuid(c.GetString("uuid"), "id")

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

	// Check if the requester is the participant of the service.
	var (
		rateDao contracts.RateDAOer
		srvDao  contracts.ServiceDAOer
	)

	depCon.Make(&rateDao)
	depCon.Make(&srvDao)

	if err := rateDao.IsServiceRatable(
		contracts.IsServiceRatableParams{
			ParticipantId: int(usr.ID),
			ServiceUuid:   srvUuid,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.ServiceNotRatable,
				err.Error(),
			),
		)

		return
	}

	srv, err := srvDao.GetServiceByUuid(body.ServiceUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return

	}

	// Create rating record.
	srvRating, err := rateDao.CreateServiceRating(
		contracts.CreateServiceRatingParams{
			Rating:      body.Rating,
			RaterId:     int(usr.ID),
			RateeId:     int(srv.GetPartnerId(usr.ID)),
			ServiceUuid: srvUuid,
			Comment:     body.Comment,
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateServiceRating,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct {
		ServiceUuid string `json:"service_uuid"`
		Rating      int32  `json:"rating"`
		Comments    string `json:"comments"`
	}{
		srvUuid,
		srvRating.Rating.Int32,
		srvRating.Comments.String,
	})
}

// CancelService allows both female and male user to cancel the service before
// service happening. Check the following conditions before
// canceling.
//   - Is a service participant.
//   - Service status is `to_be_fulfilled`.
//   - Service does not have a canceler.
// Remember to emit service canceled message to firestore.
//
// Refund mechanism:
//	 Before appointment time, both male and female can cancel the service. Male user we get full refund of matching fee.
// 	 If male cancels the service within the buffer time, darkpanda will not refund male user matching fee.
//   If female cancels the service within the buffer time, darkpanda will refund male user matching fee.
//
//   Any refund performed due to cancellation of the service, darkpanda will send out FCM message to notify the user.
func CancelService(c *gin.Context, depCon container.Container) {
	var (
		serviceUuid string = c.Param("seg")
		userUuid    string = c.GetString("uuid")
	)

	var (
		rateDao contracts.RateDAOer
		userDao contracts.UserDAOer
	)

	depCon.Make(&rateDao)
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(userUuid)

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

	isParticipant, err := rateDao.IsServiceParticipant(
		int(user.ID),
		serviceUuid,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckIsParticipant,
				err.Error(),
			),
		)

		return
	}

	if !isParticipant {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.UserNotServiceParticipant,
				err.Error(),
			),
		)

		return
	}

	var srvDao contracts.ServiceDAOer
	depCon.Make(&srvDao)

	srv, err := srvDao.GetServiceByUuid(serviceUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return
	}

	if srv.ServiceStatus != models.ServiceStatusToBeFulfilled &&
		srv.ServiceStatus != models.ServiceStatusNegotiating {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ServiceStatusNotValidToCancel),
		)

		return
	}

	if srv.CancellerID.Valid {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ServiceHasBeenCanceled),
		)

		return
	}

	var srvFsm contracts.ServiceFSMer
	depCon.Make(&srvFsm)

	srvFsm.SetState(srv.ServiceStatus.ToString())

	// @TODO "cancel" should not be hardcoded here.
	if err := srvFsm.Event("cancel"); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToChangeServiceStatus,
				err.Error(),
			),
		)

		return
	}

	var chatDao contracts.ChatDaoer
	depCon.Make(&chatDao)

	type TxResp struct {
		*models.Service
		Cause    string
		Refunded bool
	}

	trxResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			refunded := false
			cancelCause := GetCancelCause(srv.AppointmentTime.Time, user.Gender)

			if srv.ServiceStatus == models.ServiceStatusPaymentFailed {
				cancelCause = models.CancelCausePaymentFailed
			}

			log.Infof("%s cancel cause: %s", user.Uuid, string(cancelCause))

			// Change service status to cancel.
			srvStatus := models.ServiceStatus(srvFsm.Current())
			usrv, err := srvDao.WithTx(tx).UpdateServiceByID(contracts.UpdateServiceByIDParams{
				ID:            srv.ID,
				ServiceStatus: &srvStatus,
				CancellerId:   &user.ID,
				CancelCause:   &cancelCause,
			})

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToUpdateService,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			if err := chatDao.WithTx(tx).DeleteChatroomByServiceId(int(srv.ID)); err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToDeleteChatroomByServiceId,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			// Only perform refund operation when service status is `to_be_fulfilled`. This means that the service
			// is paid, male might need to refund their matching fee.
			// 2022/03/09 Note: Since DarkPanda does not charge matching fee. We don't need to perform refund anymore.
			/*
				if srv.ServiceStatus == models.ServiceStatusToBeFulfilled {
					var (
						userBalDao contracts.UserBalancer
						paymentDao contracts.PaymentDAOer
					)

					depCon.Make(&userBalDao)
					depCon.Make(&paymentDao)

					userBalDao.WithTx(tx)
					paymentDao.WithTx(tx)

					rs := NewRefundService(paymentDao, userBalDao)

					refunded, err = rs.RefundCustomerIfRefundable(srv, user)

					if err != nil {
						return db.FormatResp{
							Err:            err,
							ErrCode:        apperr.FailedToPerformRefundCustomerIfRefundable,
							HttpStatusCode: http.StatusInternalServerError,
						}
					}
				}
			*/

			return db.FormatResp{
				Response: &TxResp{
					usrv,
					string(cancelCause),
					refunded,
				},
			}
		},
	)

	if trxResp.Err != nil {
		c.AbortWithError(
			trxResp.HttpStatusCode,
			apperr.NewErr(
				trxResp.ErrCode,
				trxResp.Err.Error(),
			),
		)

		return
	}

	usrvInfo := trxResp.Response.(*TxResp)

	// Send service cancel message.
	chatroom, err := chatDao.
		WithConn(db.GetDB()).
		GetChatroomByServiceId(int(usrvInfo.ID))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChatroomByServiceId,
				err.Error(),
			),
		)
		return
	}

	ctx := context.Background()
	df := darkfirestore.Get()
	if err := df.CancelService(ctx,
		darkfirestore.CancelServiceParams{
			ChannelUuid: chatroom.ChannelUuid.String,
			ServiceUuid: usrvInfo.Uuid.String,
			Data: darkfirestore.ChatMessage{
				From: usrvInfo.Uuid.String,
			},
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendCancelMessage,
				err.Error(),
			),
		)

		return
	}

	// Emit FCM to notify service cancelled to both party.
	var dpfcmer dpfcm.DPFirebaseMessenger
	depCon.Make(&dpfcmer)

	partnerID := usrvInfo.GetPartnerId(user.ID)
	partner, err := userDao.GetUserByID(int64(partnerID), "fcm_topic")

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByID,
				err.Error(),
			),
		)

		return
	}

	if err := dpfcmer.PublishServiceCancelled(ctx, dpfcm.PublishServiceCancelledMessage{
		Topics: []string{
			user.FcmTopic.String,
			partner.FcmTopic.String,
		},
		ServiceUUID:       usrvInfo.Uuid.String,
		CancellerUUID:     user.Uuid,
		CancellerUsername: user.Username,
	}); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendServiceCancelledFCM,
				err.Error(),
			),
		)

		return
	}

	// If service is refunded, we will send refunded FCM to male user.
	if usrvInfo.Refunded {
		var maleParty *models.User

		if user.Gender == models.GenderMale {
			maleParty = user
		} else {
			maleParty = partner
		}

		if err := dpfcmer.PublishServiceRefunded(ctx, dpfcm.PublishServiceRefundedMessage{
			Topic:       maleParty.FcmTopic.String,
			ServiceUUID: usrvInfo.Uuid.String,
		}); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToSendRefundedFCM,
					err.Error(),
				),
			)

			return
		}
	}

	c.JSON(http.StatusOK, struct{}{})
}

func GetCauseWhenCancel(c *gin.Context, depCon container.Container) {
	// Check if the user is the participant of the service.
	var (
		serviceUUID string = c.Param("seg")
		userUUID    string = c.GetString("uuid")

		rateDao contracts.RateDAOer
		userDao contracts.UserDAOer
	)

	depCon.Make(&rateDao)
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(userUUID)

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

	isParticipant, err := rateDao.IsServiceParticipant(int(user.ID), serviceUUID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckIsParticipant,
				err.Error(),
			),
		)

		return
	}

	if !isParticipant {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.UserNotServiceParticipant,
				err.Error(),
			),
		)

		return
	}

	var srvDao contracts.ServiceDAOer
	depCon.Make(&srvDao)

	srv, err := srvDao.GetServiceByUuid(serviceUUID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return
	}

	cancelCause := GetCancelCause(srv.AppointmentTime.Time, user.Gender)

	c.JSON(http.StatusOK, struct {
		CancelCause string `json:"cancel_cause"`
	}{
		CancelCause: string(cancelCause),
	})
}
