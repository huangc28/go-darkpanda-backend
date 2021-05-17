package coin

import "strconv"

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
