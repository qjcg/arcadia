package blocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const URLTemplate = `https://api.kraken.com/0/public/Ticker?pair=%s` // See https://www.kraken.com/en-gb/features/api#get-ticker-info

type APIResponse struct {
	Result struct {
		Pair struct {
			P [2]string
		} `json:"XXBTZCAD"` // FIXME: generalize
	} `json:"result"`
}

type CryptoCurrency struct {
	Pair string
}

func (c *CryptoCurrency) String() string {
	url := fmt.Sprintf(URLTemplate, c.Pair)

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "http error"
	}

	resp, err := http.DefaultClient.Do(req)
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
