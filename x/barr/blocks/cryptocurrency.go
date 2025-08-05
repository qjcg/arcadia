package blocks

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// See https://www.kraken.com/en-gb/features/api#get-ticker-info
const URLTemplate = `https://api.kraken.com/0/public/Ticker?pair=%s`

type APIResponse struct {
	Result struct {
		Pair struct {
			P [2]string
		} `json:"XXBTZCAD"`
	} `json:"result"`
}

type CryptoCurrency struct {
	Pair string
}

func (c *CryptoCurrency) String() string {
	url := fmt.Sprintf(URLTemplate, c.Pair)
	resp, err := http.Get(url)
	if err != nil {
		return "http error"
	}

	defer resp.Body.Close()

	var r APIResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "json decode error"
	}
	return fmt.Sprintf("%+v", r)
}
