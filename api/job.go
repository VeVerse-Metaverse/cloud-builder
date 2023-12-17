package api

import (
	"bytes"
	"context"
	sm "dev.hackerman.me/artheon/veverse-shared/model"
	"encoding/json"
	"fmt"
	"io"
	"l7-cloud-builder/config"
	"l7-cloud-builder/logger"
	"net/http"
	"strings"
)

func FetchUnclaimedJob(ctx context.Context) (*sm.JobV2, error) {
	// Login to the API
	err := Login(ctx)
	if err != nil {
		return nil, err
	}

	// Get the list of enabled platforms
	enabledPlatforms := make([]string, 0)
	for platform, enabled := range config.Config.EnabledPlatforms {
		if enabled {
			enabledPlatforms = append(enabledPlatforms, platform)
		}
	}

	// Get the list of enabled jobs
	enabledJobs := make([]string, 0)
	for job, enabled := range config.Config.EnabledJobs {
		if enabled {
			enabledJobs = append(enabledJobs, job)
		}
	}

	// Get the list of enabled targets
	enabledTargets := make([]string, 0)
	for target, enabled := range config.Config.EnabledTargets {
		if enabled {
			enabledTargets = append(enabledTargets, target)
		}
	}

	// Prepare the request URL
	url := fmt.Sprintf("%s/job/v2/unclaimed?platform=%s&type=%s&target=%s", config.Api.Url, strings.Join(enabledPlatforms, ","), strings.Join(enabledJobs, ","), strings.Join(enabledTargets, ","))

	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, fmt.Errorf("failed to fetch unclaimed job from %s, status code: %d, error: %v", url, res.StatusCode, err)
	}

	// Prepare the response body
	var resBody []byte
	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Prepare the job container
	var c struct {
		Job     sm.JobV2 `json:"data"`
		Status  string   `json:"status"`
		Message string   `json:"message"`
	}
	err = json.Unmarshal(resBody, &c)
	if err != nil {
		return nil, err
	}

	// Handle no unclaimed jobs case
	if c.Status == "no jobs" {
		return nil, nil
	}

	// Handle error case
	if c.Status == "error" {
		return nil, fmt.Errorf("failed to fetch unclaimed job from %s, status code: %d, error: %v", url, res.StatusCode, c.Message)
	}

	// Return the job
	return &c.Job, nil
}

func UpdateJobStatus(ctx context.Context, job *sm.JobV2, status config.JobStatusType, message string) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}

	// Login to the API
	err := Login(ctx)
	if err != nil {
		return err
	}

	// Prepare the request URL
	url := fmt.Sprintf("%s/job/v2/%s/status", config.Api.Url, job.Id)

	// Prepare the request body
	body := struct {
		Status  config.JobStatusType `json:"status"`
		Message string               `json:"message"`
	}{
		Status:  status,
		Message: message,
	}

	// Marshal the body
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Prepare the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(bodyBytes))
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

	// Check the response status code
	if res.StatusCode >= 400 {
		return fmt.Errorf("failed to update job status, status code: %d, error: %v", res.StatusCode, err)
	}

	return nil
}
