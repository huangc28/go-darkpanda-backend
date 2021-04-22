package image

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type ImageHandlers struct {
	Container container.Container
}

// append timestamp on each file name
// check mime types
func UploadAvatarHandler(c *gin.Context) {
	// ------------------- initialize gcs storage client -------------------
	gcsCreds := config.GetAppConf().GCSCredentials

	ctx := context.Background()
	client, err := storage.NewClient(
		ctx,
		option.WithServiceAccountFile(fmt.Sprintf("%s/%s", config.GetProjRootPath(), gcsCreds.GoogleServiceAccountName)),
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToInitGCSClient,
				err.Error(),
			),
		)

		return
	}

	enhancer := NewGCSEnhancer(client, gcsCreds.BucketName)

	// ------------------- retrieve file from multipart -------------------
	file, handler, err := c.Request.FormFile("image")

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToRetrieveFormFileFromRequest,
				err.Error(),
			),
		)

		return
	}

	pubLink, err := enhancer.Upload(ctx, file, handler.Filename)

	if err != nil {
		log.Debug("Failed to upload file to GCS %s", err.Error())

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCopyFileToGCS,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct {
		PublicLink string `json:"public_link"`
	}{
		pubLink,
	})
}

func (h *ImageHandlers) UploadImagesHandler(c *gin.Context, depCon container.Container) {
	// ------------------- Limit upload size to 20 MB -------------------
	if err := c.Request.ParseMultipartForm(20e6); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(

				apperr.FailedToParseMultipartForm,
				err.Error(),
			),
		)

		return
	}

	fileHeaders := c.Request.MultipartForm.File["image"]

	uuid := c.Request.MultipartForm.Value["uuid"]
	age := c.Request.MultipartForm.Value["age"]
	height := c.Request.MultipartForm.Value["height"]
	weight := c.Request.MultipartForm.Value["weight"]
	description := c.Request.MultipartForm.Value["description"]

	ageInt, err := strconv.Atoi(age[0])
	heightFloat, err := strconv.ParseFloat(height[0], 32)
	weightFloat, err := strconv.ParseFloat(weight[0], 32)

	gcsCreds := config.GetAppConf().GCSCredentials

	ctx := context.Background()
	client, err := storage.NewClient(
		ctx,
		option.WithServiceAccountFile(fmt.Sprintf("%s/%s", config.GetProjRootPath(), gcsCreds.GoogleServiceAccountName)),
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToInitGCSClient,
				err.Error(),
			),
		)

		return
	}

	enhancer := NewGCSEnhancer(client, gcsCreds.BucketName)
	linkList, err := enhancer.UploadMultiple(ctx, fileHeaders)

	if err != nil {
		log.Fatalf("Failed to upload multiple files %s", err.Error())
	}

	var userDAO contracts.UserDAOer
	h.Container.Make(&userDAO)

	user, err := userDAO.UpdateUserInfoByUuid(contracts.UpdateUserInfoParams{
		Age:         &ageInt,
		Height:      &heightFloat,
		Weight:      &weightFloat,
		Description: &description[0],
		Uuid:        uuid[0],
	})

	if len(linkList) > 0 {
		dao := NewImageDAO(db.GetDB())
		imagesParams := make([]CreateImageParams, 0)
		for i := 0; i < len(linkList); i++ {
			imagesParams = append(imagesParams, CreateImageParams{
				UserID: user.ID,
				URL:    linkList[i],
			})
		}

		if err := dao.CreateImages(imagesParams); err != nil {
			log.Fatalf("Failed to insert images %s", err.Error())
		}
	}

	c.JSON(http.StatusOK, NewTransformer().TransformLinks(linkList))
}
