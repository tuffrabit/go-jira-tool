// Copyright 2026 Robert Boucher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type JiraClient struct {
	config     *Config
	httpClient *http.Client
	authHeader string
}

func NewJiraClient(cfg *Config) *JiraClient {
	creds := cfg.Email + ":" + cfg.APIToken
	encoded := base64.StdEncoding.EncodeToString([]byte(creds))
	return &JiraClient{
		config:     cfg,
		httpClient: &http.Client{},
		authHeader: "Basic " + encoded,
	}
}

func (c *JiraClient) apiURL(path string) string {
	return strings.TrimRight(c.config.BaseURL, "/") + "/rest/api/3" + path
}

func (c *JiraClient) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

func (c *JiraClient) Get(path string) (*http.Response, error) {
	return c.doRequest("GET", c.apiURL(path), nil)
}

func (c *JiraClient) Post(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.doRequest("POST", c.apiURL(path), strings.NewReader(string(data)))
}

func (c *JiraClient) Put(path string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.doRequest("PUT", c.apiURL(path), strings.NewReader(string(data)))
}

// GetRaw makes an authenticated GET request to an arbitrary URL (e.g. attachment download URLs).
func (c *JiraClient) GetRaw(url string) (*http.Response, error) {
	return c.doRequest("GET", url, nil)
}

// ReadJSONBody reads and decodes a JSON response body, closing it afterward.
// If the response status is not 2xx, it returns an error with the status and body.
func ReadJSONBody(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
	if target != nil {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to parse response JSON: %w", err)
		}
	}
	return nil
}
