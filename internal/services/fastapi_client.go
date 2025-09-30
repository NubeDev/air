package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/NubeDev/air/internal/logger"
	"github.com/go-resty/resty/v2"
)

// FastAPIClient handles communication with the FastAPI data processing service
type FastAPIClient struct {
	client  *resty.Client
	baseURL string
}

// FastAPI request/response models
type DiscoverRequest struct {
	DatasourceID string `json:"datasource_id"`
	URI          string `json:"uri"`
	Recurse      bool   `json:"recurse"`
	MaxFiles     *int   `json:"max_files,omitempty"`
}

type InferSchemaRequest struct {
	DatasourceID string `json:"datasource_id"`
	URI          string `json:"uri"`
	SampleFiles  *int   `json:"sample_files,omitempty"`
	InferRows    int    `json:"infer_rows"`
}

type PreviewRequest struct {
	DatasourceID string  `json:"datasource_id"`
	Path         *string `json:"path,omitempty"`
	Limit        int     `json:"limit"`
}

type AnalyzeRequest struct {
	DatasourceID *string     `json:"datasource_id,omitempty"`
	Plan         *QueryPlan  `json:"plan,omitempty"`
	JobKind      string      `json:"job_kind"`
	Options      interface{} `json:"options,omitempty"`
}

type QueryPlan struct {
	Dataset string                   `json:"dataset"`
	Select  []string                 `json:"select,omitempty"`
	Filters []map[string]interface{} `json:"filters,omitempty"`
}

type TokenResponse struct {
	Token int `json:"token"`
}

type JobStatusResponse struct {
	Token  int         `json:"token"`
	Status string      `json:"status"`
	Steps  []JobStep   `json:"steps"`
	Data   interface{} `json:"data,omitempty"`
	Code   int         `json:"code"`
	Error  *string     `json:"error,omitempty"`
}

type JobStep struct {
	Step      int    `json:"step"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Duration  *int   `json:"duration_ms,omitempty"`
}

// NewFastAPIClient creates a new FastAPI client
func NewFastAPIClient(baseURL string) *FastAPIClient {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)

	return &FastAPIClient{
		client:  client,
		baseURL: baseURL,
	}
}

// Health checks if the FastAPI service is healthy
func (c *FastAPIClient) Health() error {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get(c.baseURL + "/v1/py/health")

	if err != nil {
		logger.LogError(logger.ServiceAI, "FastAPI health check failed", err)
		return err
	}

	if resp.StatusCode() != 200 {
		logger.LogError(logger.ServiceAI, "FastAPI health check failed", fmt.Errorf("status code: %d", resp.StatusCode()))
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode())
	}

	logger.LogInfo(logger.ServiceAI, "FastAPI service is healthy")
	return nil
}

// DiscoverFiles discovers files in a directory
func (c *FastAPIClient) DiscoverFiles(req DiscoverRequest) (*TokenResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/v1/py/discover")

	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to discover files", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		logger.LogError(logger.ServiceAI, "Discover files failed", fmt.Errorf("status code: %d", resp.StatusCode()))
		return nil, fmt.Errorf("discover files failed with status: %d", resp.StatusCode())
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse discover response", err)
		return nil, err
	}

	logger.LogInfo(logger.ServiceAI, "File discovery started", map[string]interface{}{
		"token": tokenResp.Token,
		"uri":   req.URI,
	})

	return &tokenResp, nil
}

// InferSchema infers schema from files
func (c *FastAPIClient) InferSchema(req InferSchemaRequest) (*TokenResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/v1/py/infer_schema")

	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to infer schema", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		logger.LogError(logger.ServiceAI, "Infer schema failed", fmt.Errorf("status code: %d", resp.StatusCode()))
		return nil, fmt.Errorf("infer schema failed with status: %d", resp.StatusCode())
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse infer schema response", err)
		return nil, err
	}

	logger.LogInfo(logger.ServiceAI, "Schema inference started", map[string]interface{}{
		"token": tokenResp.Token,
		"uri":   req.URI,
	})

	return &tokenResp, nil
}

// PreviewData previews data from files
func (c *FastAPIClient) PreviewData(req PreviewRequest) (*TokenResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/v1/py/preview")

	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to preview data", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		logger.LogError(logger.ServiceAI, "Preview data failed", fmt.Errorf("status code: %d", resp.StatusCode()))
		return nil, fmt.Errorf("preview data failed with status: %d", resp.StatusCode())
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse preview response", err)
		return nil, err
	}

	logger.LogInfo(logger.ServiceAI, "Data preview started", map[string]interface{}{
		"token": tokenResp.Token,
		"path":  req.Path,
	})

	return &tokenResp, nil
}

// AnalyzeData analyzes data for EDA, profiling, validation
func (c *FastAPIClient) AnalyzeData(req AnalyzeRequest) (*TokenResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.baseURL + "/v1/py/analyze")

	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to analyze data", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		logger.LogError(logger.ServiceAI, "Analyze data failed", fmt.Errorf("status code: %d", resp.StatusCode()))
		return nil, fmt.Errorf("analyze data failed with status: %d", resp.StatusCode())
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse analyze response", err)
		return nil, err
	}

	logger.LogInfo(logger.ServiceAI, "Data analysis started", map[string]interface{}{
		"token":    tokenResp.Token,
		"job_kind": req.JobKind,
	})

	return &tokenResp, nil
}

// GetJobStatus gets the status of a job
func (c *FastAPIClient) GetJobStatus(token int) (*JobStatusResponse, error) {
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/v1/py/jobs/%d", c.baseURL, token))

	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to get job status", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		logger.LogError(logger.ServiceAI, "Get job status failed", fmt.Errorf("status code: %d", resp.StatusCode()))
		return nil, fmt.Errorf("get job status failed with status: %d", resp.StatusCode())
	}

	var jobResp JobStatusResponse
	if err := json.Unmarshal(resp.Body(), &jobResp); err != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse job status response", err)
		return nil, err
	}

	return &jobResp, nil
}

// WaitForJobCompletion waits for a job to complete and returns the final result
func (c *FastAPIClient) WaitForJobCompletion(token int, maxWaitTime time.Duration) (*JobStatusResponse, error) {
	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := c.GetJobStatus(token)
			if err != nil {
				return nil, err
			}

			if status.Status == "completed" || status.Status == "failed" {
				logger.LogInfo(logger.ServiceAI, "Job completed", map[string]interface{}{
					"token":  token,
					"status": status.Status,
					"steps":  len(status.Steps),
				})
				return status, nil
			}

			if time.Since(start) > maxWaitTime {
				return nil, fmt.Errorf("job %d timed out after %v", token, maxWaitTime)
			}

			logger.LogInfo(logger.ServiceAI, "Job in progress", map[string]interface{}{
				"token":  token,
				"status": status.Status,
				"step":   len(status.Steps),
			})
		}
	}
}
