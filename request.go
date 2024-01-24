package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func getRawFromHttpReq(url string, headers map[string]string) ([]byte, error) {
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

func getJsonFromHttpReq[TBody any](url string, headers map[string]string) (*TBody, error) {
	body, err := getRawFromHttpReq(url, headers)
	if err != nil {
		return nil, err
	}
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
