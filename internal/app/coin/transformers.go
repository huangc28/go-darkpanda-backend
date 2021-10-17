package coin

import (
	"strconv"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/shopspring/decimal"
)

type TransformedCoinBalance struct {
	Balance float64 `json:"balance"`
}

func TransformBuyCoin(bal string) (TransformedCoinBalance, error) {
	balFloat, err := strconv.ParseFloat(bal, 32)

	if err != nil {
		return TransformedCoinBalance{}, err
	}

	return TransformedCoinBalance{
		Balance: balFloat,
	}, nil
}

func TransformGetCoinBalance(bal string) (TransformedCoinBalance, error) {
	balFloat, err := strconv.ParseFloat(bal, 32)

	if err != nil {
		return TransformedCoinBalance{}, err
	}

	return TransformedCoinBalance{
		Balance: balFloat,
	}, nil
}

type TransformedPkg struct {
	Id      int     `json:"id"`
	DPCoins int     `json:"dp_coin"`
	Cost    float64 `json:"cost"`
}

type TransformedPkgs struct {
	Packages []TransformedPkg `json:"packages"`
}

func TransformCoinPakcages(pkgModels []models.CoinPackage) (TransformedPkgs, error) {
	trfms := make([]TransformedPkg, 0)

	trfPkgs := TransformedPkgs{}

	for _, m := range pkgModels {
		costDeci, err := decimal.NewFromString(m.Cost.String)

		if err != nil {
			return trfPkgs, err
		}

		costF, _ := costDeci.Float64()

		trfm := TransformedPkg{
			Id:      int(m.ID),
			DPCoins: int(m.DbCoins.Int32),
			Cost:    costF,
		}

		trfms = append(trfms, trfm)

	}

	trfPkgs.Packages = trfms

	return trfPkgs, nil
}
