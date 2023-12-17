package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"l7-cloud-builder/config"
	"l7-cloud-builder/logger"
	"net/http"
)

func Login(ctx context.Context) (err error) {
	// Prepare the request body
	var reqBody []byte
	reqBody, err = json.Marshal(map[string]string{
		"email":    config.Api.Email,
		"password": config.Api.Password,
	})
	if err != nil {
		return err
	}

	// Prepare the request URL
	url := fmt.Sprintf("%s/auth/login", config.Api.Url)

	// Prepare the request
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")

	// Prepare the client
	var client = &http.Client{}

	// Prepare the response
	var res *http.Response

	// Send the request
	res, err = client.Do(req)
	if err != nil {
		return err
	}

	// Defer closing the response body
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			logger.Logger.Warningf("error closing http response body: %v\n", err)
		}
	}(res.Body)

	// Check the response status code
	if res.StatusCode >= 400 {
		return fmt.Errorf("failed to login to %s, status code: %d, error: %v", url, res.StatusCode, err)
	}

	// Read the response body
	var body []byte
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Unmarshal the response body
	var v map[string]string
	if err = json.Unmarshal(body, &v); err != nil {
		return err
	}

	// Check if the map contains the data key
	if _, ok := v["data"]; ok {
		// Use the data key as the token
		config.Api.Token = v["data"]
		return nil
	} else if _, ok := v["message"]; ok {
		// Return an error with the message
		return fmt.Errorf("failed to login to %s, status code: %d, error: %s", url, res.StatusCode, v["message"])
	} else {
		// Return an error with no message
		return fmt.Errorf("failed to login to %s, status code: %d, no error message in response", url, res.StatusCode)
	}
}
