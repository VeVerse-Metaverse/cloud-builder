package api

import (
	"context"
	sm "dev.hackerman.me/artheon/veverse-shared/model"
	"encoding/json"
	"fmt"
	"io"
	"l7-cloud-builder/config"
	"l7-cloud-builder/logger"
	"net/http"
)

func LoadSharedConfiguration(ctx context.Context) error {
	// Login to the API
	err := Login(ctx)
	if err != nil {
		return err
	}

	// Prepare the request URL
	url := fmt.Sprintf("%s/automation/configuration", config.Api.Url)

	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Api.Token))

	// Prepare the client
	client := &http.Client{}

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
		return fmt.Errorf("failed to fetch unclaimed job from %s, status code: %d, error: %v", url, res.StatusCode, err)
	}

	// Prepare the response body
	var resBody []byte
	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Prepare the job container
	var c struct {
		Configuration sm.AutomationConfiguration `json:"data"` // todo: implement this
		Status        string                     `json:"status"`
		Message       string                     `json:"message"`
	}
	err = json.Unmarshal(resBody, &c)
	if err != nil {
		return err
	}

	// Handle error case
	if c.Status == "error" {
		return fmt.Errorf("failed to fetch unclaimed job from %s, status code: %d, error: %v", url, res.StatusCode, c.Message)
	}

	config.Shared.Release.IgnoredFiles = c.Configuration.Release.IgnoredFiles // todo: implement this

	// Return the job
	return nil
}
