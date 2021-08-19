package convertnullsql

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

func ConvertFloatStringToDecimal(floatStr sql.NullString) (decimal.Decimal, error) {
	if !floatStr.Valid {
		return decimal.Decimal{}, nil
	}

	return decimal.NewFromString(floatStr.String)
}

func ConvertSqlNullStringToFloat64(floatStr sql.NullString) (*float64, error) {
	dec, err := ConvertFloatStringToDecimal(floatStr)

	if err != nil {
		return nil, err
	}

	floatNum, _ := dec.Float64()

	return &floatNum, nil
}

func ConvertSqlNullStringToFloat32(floatStr sql.NullString) (*float32, error) {
	dec, err := ConvertFloatStringToDecimal(floatStr)

	if err != nil {
		return nil, err
	}

	floatNum, _ := dec.BigFloat().Float32()

	return &floatNum, nil
}
