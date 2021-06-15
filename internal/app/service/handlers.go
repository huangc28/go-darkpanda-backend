package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
)

// GetListOfCurrentServicesHandler retrieve those service of the following status:
//   - unpaid
//   - to_be_fulfilled
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

	srvs, err = srvDao.GetServicesByStatus(
		int(user.ID),
		user.Gender,
		body.Offset,
		body.PerPage,
		models.ServiceStatusUnpaid,
		models.ServiceStatusToBeFulfilled,
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

	log.Printf("DEBUG 1 channel uuid  %v %v", channelUuids, len(channelUuids))

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
	// We have service `appointment time ` and `service duration`
	err = IsTimeInRange(
		srv.AppointmentTime.Time,
		srv.AppointmentTime.Time.Add(4*time.Hour),
	)

	df := darkfirestore.Get()
	ctx := context.Background()

	srvStatusExp := models.ServiceStatusExpired
	if errors.Is(err, ErrorExpired) {
		// Update service status in firestore.
		df.UpdateService(
			ctx,
			darkfirestore.UpdateServiceParams{
				ServiceUuid:   srv.Uuid.String(),
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

			err = df.UpdateService(ctx, darkfirestore.UpdateServiceParams{
				ServiceUuid:   usrv.Uuid.String(),
				ServiceStatus: usrv.ServiceStatus.ToString(),
			})

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FirestoreFailedToUpdateService,
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

	srvNames, err := srvDao.GetServiceNames()

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

	c.JSON(http.StatusOK, TransformServiceName(srvNames))
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

	if srv.IsNotOneOfStatus(
		models.ServiceStatusExpired,
		models.ServiceStatusCompleted,
	) {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.ServiceNotYetEnd),
		)

		return
	}

	var pDaoer contracts.PaymentDAOer
	depCon.Make(&pDaoer)

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

	c.JSON(http.StatusOK, TrfPaymentDetail(p))
}
