package coin

import "strconv"

type TransformedBuyCoin struct {
	Balance float64 `json:"balance"`
}

func TransformBuyCoin(bal string) (TransformedBuyCoin, error) {
	balFloat, err := strconv.ParseFloat(bal, 32)

	if err != nil {
		return TransformedBuyCoin{}, err

	}

	return TransformedBuyCoin{
		Balance: balFloat,
	}, nil
}
