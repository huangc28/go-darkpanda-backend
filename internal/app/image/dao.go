package image

import (
	"database/sql"
	"strings"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type ImageDAO struct {
	DB *sql.DB
}

type ImageDAOer interface {
	GetImagesByUserID(ID int) []models.Image
}

func (dao *ImageDAO) GetImagesByUserID(ID int) ([]models.Image, error) {
	sql := `
SELECT url
FROM images
WHERE user_id = $1
	`
	rows, err := dao.DB.Query(sql, ID)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	images := make([]models.Image, 0)

	for rows.Next() {
		var image models.Image
		if err := rows.Scan(&image); err != nil {
			return nil, err
		}

		images = append(images, image)
	}

	return images, nil
}

type CreateImageParams struct {
	UserID int64
	URL    string
}

func (dao *ImageDAO) CreateImages(imagesParams []CreateImageParams) error {
	sqlStr := "INSERT INTO images(user_id, url) VALUES "
	vals := []interface{}{}

	for _, v := range imagesParams {
		sqlStr += "(?, ?),"
		vals = append(vals, v.UserID, v.URL)
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")
	pgStr := db.ReplaceSQLPlaceHolderWithPG(sqlStr, "?")

	stmt, _ := dao.DB.Prepare(pgStr)
	_, err := stmt.Exec(vals...)

	if err != nil {
		return err
	}

	return nil
}
