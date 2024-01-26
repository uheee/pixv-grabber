package request

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func GetRawFromHttpReq(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetJsonFromHttpReq[TBody any](url string, headers map[string]string) (*TBody, error) {
	body, err := GetRawFromHttpReq(url, headers)
	if err != nil {
		return nil, err
	}
	j := string(body[:])
	log.Debug().Str("json", j).Msg("get json from http req")
	var result Response[TBody]
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.Error {
		return nil, errors.New(result.Message)
	}
	return &result.Body, nil
}
