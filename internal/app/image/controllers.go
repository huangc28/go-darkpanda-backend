package image

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

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

func UploadImagesHandler(c *gin.Context) {
	// ------------------- Limit upload size to 20 MB -------------------
	if err := c.Request.ParseMultipartForm(20E6); err != nil {
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

	c.JSON(http.StatusOK, NewTransformer().TransformLinks(linkList))
}
