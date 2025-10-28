package http_request

import (
	"bytes"
	"encoding/json"
	"go-api-boilerplate/module/logger"
	"io"
	"net/http"
)

func HttpRequest(url string, method string, headers map[string]string, queries map[string]string, body map[string]any) (string, int, error) {
	requestBody := &bytes.Buffer{}
	result := ""
	statusCode := 0

	logger.Log.Infof("url: %s, method: %s, headers: %v, queries: %v, body: %v", url, method, headers, queries, body)

	switch method {
	case http.MethodPost:
		if len(body) > 0 {
			postBody, err := json.Marshal(body)
			if err != nil {
				return result, statusCode, err
			}

			requestBody = bytes.NewBuffer(postBody)
		}
	}

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		logger.Log.Errorf("url: %s, error: %s", url, err)
		return result, statusCode, ErrNewRequest
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	qry := req.URL.Query()

	for key, value := range queries {
		qry.Add(key, value)
	}

	if len(queries) > 0 {
		req.URL.RawQuery = qry.Encode()
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Errorf("url: %s, error: %s", url, err)
		return result, statusCode, ErrDoRequest
	}

	defer response.Body.Close()
	statusCode = response.StatusCode

	output, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Errorf("url: %s, error: %s", url, err)
		return result, statusCode, ErrIOReadResponse
	}

	result = string(output)

	logger.Log.Infof("url: %s, response: %s", url, result)

	return result, statusCode, nil
}
