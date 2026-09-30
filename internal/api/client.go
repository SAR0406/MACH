package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTP: &http.Client{}}
}

func (c *Client) RegisterChallenge(req RegisterChallengeRequest) (RegisterChallengeResponse, error) {
	var resp RegisterChallengeResponse
	err := c.request(http.MethodPost, "/v1/register/challenge", "", req, &resp)
	return resp, err
}

func (c *Client) RegisterComplete(req RegisterCompleteRequest) (RegisterCompleteResponse, error) {
	var resp RegisterCompleteResponse
	err := c.request(http.MethodPost, "/v1/register/complete", "", req, &resp)
	return resp, err
}

func (c *Client) Heartbeat(req HeartbeatRequest) error {
	return c.request(http.MethodPost, "/v1/device/heartbeat", c.Token, req, nil)
}

func (c *Client) Expose(port int) (Exposure, error) {
	var ex Exposure
	err := c.request(http.MethodPost, "/v1/device/expose", c.Token, ExposeRequest{Port: port}, &ex)
	return ex, err
}

func (c *Client) Status() (DeviceStatusResponse, error) {
	var out DeviceStatusResponse
	err := c.request(http.MethodGet, "/v1/device/status", c.Token, nil, &out)
	return out, err
}

func (c *Client) Stop(exposureID string) error {
	return c.request(http.MethodPost, "/v1/device/stop", c.Token, StopRequest{ExposureID: exposureID}, nil)
}

func (c *Client) Logs(limit int) (LogsResponse, error) {
	var out LogsResponse
	path := "/v1/device/logs"
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}
	err := c.request(http.MethodGet, path, c.Token, nil, &out)
	return out, err
}

func (c *Client) request(method, path, token string, body any, out any) error {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
