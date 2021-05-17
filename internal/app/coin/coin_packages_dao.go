package coin

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type CoinPackagesDAO struct {
	db db.Conn
}

func NewCoinPackagesDAO(db db.Conn) *CoinPackagesDAO {
	return &CoinPackagesDAO{
		db: db,
	}
}

func (dao *CoinPackagesDAO) GetPackages() ([]models.CoinPackage, error) {
	query := `
SELECT * FROM coin_packages;
	`

	rows, err := dao.db.Queryx(query)

	pkgs := make([]models.CoinPackage, 0)

	if err != nil {
		return pkgs, err
	}

	for rows.Next() {
		pkg := models.CoinPackage{}

		if err := rows.StructScan(&pkg); err != nil {
			return pkgs, err
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func (dao *CoinPackagesDAO) GetPackageById(Id int) (*models.CoinPackage, error) {
	query := `
SELECT *
FROM coin_packages
WHERE coin_packages.id = $1;
	`

	pkg := models.CoinPackage{}

	if err := dao.db.QueryRowx(query, Id).StructScan(&pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}
