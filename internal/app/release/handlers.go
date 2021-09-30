package release

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
)

type AppCenterLatestReleaseResponse struct {
	DownloadURL string `json:"download_url"`
	InstallURL  string `json:"install_url"`
}

func AndroidLatestDLLinkHandler(c *gin.Context) {
	conf := config.GetAppConf()

	// Request appcenter openapi.
	url := url.URL{
		Scheme: "https",
		Host:   "api.appcenter.ms",
		Path: fmt.Sprintf(
			"v0.1/public/sdk/apps/%s/distribution_groups/%s/releases/latest",
			conf.AppcenterAppSecret,
			conf.AppcenterPublicDistributionGroupId,
		),
	}

	r, err := http.NewRequest("GET", url.String(), nil)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToInitAppcenterRequest,
				err.Error(),
			),
		)

		return
	}

	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendAppcenterOpenApiRequest,
				err.Error(),
			),
		)

		return
	}

	defer resp.Body.Close()

	ar := &AppCenterLatestReleaseResponse{}
	dec := json.NewDecoder(resp.Body)

	if err := dec.Decode(ar); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToScanAppcenterResponse,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, ar)
}
