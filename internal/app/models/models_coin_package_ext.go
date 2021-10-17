package models

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func (cp *CoinPackage) MatchingFeeRate() (float64, error) {
	if !cp.Name.Valid || cp.Name.String != "matching_fee_rate" {
		return 0, fmt.Errorf("current instance name is not matching_fee_rate")
	}

	rateDeci, err := decimal.NewFromString(cp.Cost.String)

	if err != nil {
		return 0, err
	}

	rateF, _ := rateDeci.Float64()

	return rateF, nil
}

func (cp *CoinPackage) CalcMatchingFee(price float64) (float64, error) {
	mr, err := cp.MatchingFeeRate()

	if err != nil {
		return 0, err
	}

	return mr * price, nil
}

func (cp *CoinPackage) CalcMatchingFeeFromString(price string) (float64, error) {
	pDeci, err := decimal.NewFromString(price)

	if err != nil {
		return 0, err
	}

	pF, _ := pDeci.Float64()

	return cp.CalcMatchingFee(pF)
}
