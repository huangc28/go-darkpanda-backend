package coin

import (
	"strconv"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
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
	Id      int `json:"id"`
	DPCoins int `json:"dp_coin"`
	Cost    int `json:"cost"`
}

type TransformedPkgs struct {
	Packages []TransformedPkg `json:"packages"`
}

func TransformCoinPakcages(pkgModels []models.CoinPackage) TransformedPkgs {
	trfms := make([]TransformedPkg, 0)

	for _, m := range pkgModels {
		trfm := TransformedPkg{
			Id:      int(m.ID),
			DPCoins: int(m.DbCoins.Int32),
			Cost:    int(m.Cost.Int32),
		}

		trfms = append(trfms, trfm)

	}

	trfPkgs := TransformedPkgs{
		Packages: trfms,
	}

	return trfPkgs
}
