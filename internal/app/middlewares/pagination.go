package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
)

type PaginateParams struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"per_page,default=10"`
}

type HasMoreQuerier interface {
	HasMore(offset, perPage int) (bool, error)
}

func Pagination(querier HasMoreQuerier) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := PaginateParams{}

		// Parse pagination from query.
		if err := c.ShouldBindQuery(&params); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToParsePaginateParams,
					err.Error(),
				),
			)

			return

		}

		c.Set("offset", params.Offset)
		c.Set("per_page", params.PerPage)

		c.Next()

		// Execute query to retrieve `has_more` info.
		querier.HasMore(
			params.Offset,
			params.PerPage,
		)

		//c.JSON()
	}
}
