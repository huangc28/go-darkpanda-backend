package image

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	gcsenhancer "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/gcs_enhancer"

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
	appConf := config.GetAppConf()

	ctx := context.Background()
	client, err := storage.NewClient(
		ctx,
		option.WithCredentialsFile(
			fmt.Sprintf("%s/%s", config.GetProjRootPath(),
				appConf.GcsGoogleServiceAccountName,
			),
		),
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

	enhancer := gcsenhancer.NewGCSEnhancer(
		client,
		appConf.GcsBucketName,
	)

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

// @TODOs
//   - Accept only png and jpg.
//   - We need to determine the appropriate mime type to use proper decoder.
func UploadImagesHandler(c *gin.Context, depCon container.Container) {
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

	// sis, err := CropThumbnail(c.Request.MultipartForm.File["image"])
	cis, err := CompressImages(c.Request.MultipartForm.File["image"])

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCropImages,
				err.Error(),
			),
		)

		return
	}

	appConf := config.GetAppConf()
	ctx := context.Background()
	client, err := storage.NewClient(
		ctx,
		option.WithCredentialsFile(
			fmt.Sprintf(
				"%s/%s",
				config.GetProjRootPath(),
				appConf.GcsGoogleServiceAccountName,
			),
		),
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

	enhancer := gcsenhancer.NewGCSEnhancer(
		client,
		appConf.GcsBucketName,
	)

	imgs := make([]gcsenhancer.Images, 0)

	for _, si := range cis {
		imgs = append(imgs, gcsenhancer.Images{
			Name:      si.Name,
			Mime:      si.Mime,
			OrigImage: si.OrigImage,
			Thumbnail: si.CompressedImage,
		})
	}

	sl, err := enhancer.UploadImages(ctx, imgs)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUploadImagesToGCS,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, sl)
}
