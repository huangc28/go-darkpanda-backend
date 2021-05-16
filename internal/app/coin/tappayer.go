package coin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type TapPayerConf struct {
	Url        string
	PartnerKey string `json:"partner_key"`
	MerchantId string `json:"merchant_id"`
}

type TapPayer struct {
	conf TapPayerConf
}

func NewTapPayer(conf TapPayerConf) *TapPayer {
	return &TapPayer{
		conf: conf,
	}
}

type CardHolderParams struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	ZipCode     string `json:"zip_code"`
	Address     string `json:"address"`
	NationalId  string `json:"national_id"`
}

type PayByPrimeParams struct {
	TapPayerConf
	Prime      string           `json:"prime"`
	Details    string           `json:"details"`
	Amount     string           `json:"amount"`
	Currency   string           `json:"currency"`
	Cardholder CardHolderParams `json:"cardholder"`
	Remember   bool             `json:"remember"`
}

// Tappay API response code.
type TapPayResponseStatus int

var (
	PayByPrimeOk = 0
)

type TapPayResponse struct {
	Status            int     `json:"status"`
	Msg               string  `json:"msg"`
	RecTradeId        string  `json:"rec_trade_id"`
	BankTransactionId string  `json:"bank_transaction_id"`
	AuthCode          string  `json:"auth_code"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Raw               string
}

func (t *TapPayer) PayByPrime(params PayByPrimeParams) (*TapPayResponse, string, error) {
	params.TapPayerConf = t.conf

	buf, err := json.Marshal(params)

	if err != nil {
		return nil, "", err
	}

	req, _ := http.NewRequest(
		"POST",
		t.conf.Url,
		bytes.NewBuffer(buf),
	)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", t.conf.PartnerKey)

	client := &http.Client{}
	resp, err := client.Do(req)

	respByte, err := ioutil.ReadAll(resp.Body)
	respStr := string(respByte)

	// Marsh tp response to json string.
	if err != nil {
		return nil, respStr, err
	}

	defer resp.Body.Close()

	tpResp := TapPayResponse{}

	if err := json.Unmarshal(respByte, &tpResp); err != nil {
		return nil, respStr, err
	}

	// If tappay returns error response, we compose an error here.
	if tpResp.Status != PayByPrimeOk {
		return nil, respStr, errors.New(tpResp.Msg)
	}

	return &tpResp, respStr, nil
}
