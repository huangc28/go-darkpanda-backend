package models

import (
	"fmt"
	"math"
	"strconv"

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

	if rateF == 0.0 {
		return 0, fmt.Errorf("matching fee rate can not be 0.0, makesure the value is properly set in db")

	}

	return rateF, nil
}

func (cp *CoinPackage) CalcMatchingFee(price float64) (float64, error) {
	mr, err := cp.MatchingFeeRate()

	if err != nil {
		return 0, err
	}

	// Round matching fee to nearest 2 precision.
	return math.Round(mr*price*100) / 100, nil
}

func (cp *CoinPackage) CalcMatchingFeeFromString(price string) (float64, error) {
	pDeci, err := decimal.NewFromString(price)

	if err != nil {
		return 0, err
	}

	pF, _ := pDeci.Float64()

	return cp.CalcMatchingFee(pF)
}

func (cp *CoinPackage) IntCost() (int, error) {
	fAmount, err := strconv.ParseFloat(cp.Cost.String, 64)

	if err != nil {
		return 0, err
	}

	return int(fAmount), nil
}
