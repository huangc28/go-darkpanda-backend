package coin

import (
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type PackageName string

const (
	DpCoin200   PackageName = "dp_coin_200"
	DpCoin500               = "dp_coin_500"
	DpCoin1000              = "dp_coin_1000"
	DpCoin2000              = "dp_coin_2000"
	MatchingFee             = "matching_fee"
)

type CoinPackagesDAO struct {
	db db.Conn
}

func NewCoinPackagesDAO(db db.Conn) *CoinPackagesDAO {
	return &CoinPackagesDAO{
		db: db,
	}
}

func CoinPackageDaoServiceProvider(c cintrnal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.CoinPackageDAOer {
			return NewCoinPackagesDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *CoinPackagesDAO) GetPackages() ([]models.CoinPackage, error) {
	query := `
SELECT
	*
FROM
	coin_packages
WHERE
	name <> $1;
	`

	rows, err := dao.
		db.
		Queryx(
			query,
			MatchingFee,
		)

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

func (dao *CoinPackagesDAO) GetMatchingFee() (*models.CoinPackage, error) {
	query := `
SELECT *
FROM coin_packages
WHERE name = $1;
`

	var m models.CoinPackage

	if err := dao.db.QueryRowx(query, MatchingFee).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}
