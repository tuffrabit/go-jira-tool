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
	"encoding/json"
	"fmt"
	"net/url"
)

type SearchResult struct {
	Key     string   `json:"key"`
	Summary string   `json:"summary"`
	Status  string   `json:"status"`
	Labels  []string `json:"labels"`
}

type searchResponse struct {
	Issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
			Status  struct {
				Name string `json:"name"`
			} `json:"status"`
			Labels []string `json:"labels"`
		} `json:"fields"`
	} `json:"issues"`
	NextPageToken string `json:"nextPageToken"`
}

func SearchIssues(client *JiraClient, jql string) ([]SearchResult, error) {
	var allResults []SearchResult
	nextPageToken := ""

	for {
		params := url.Values{}
		params.Set("jql", jql)
		params.Set("fields", "summary,status,labels")
		params.Set("maxResults", "50")
		if nextPageToken != "" {
			params.Set("nextPageToken", nextPageToken)
		}

		resp, err := client.Get("/search/jql?" + params.Encode())
		if err != nil {
			return nil, fmt.Errorf("search request failed: %w", err)
		}

		var searchResp searchResponse
		if err := ReadJSONBody(resp, &searchResp); err != nil {
			return nil, fmt.Errorf("failed to read search response: %w", err)
		}

		for _, issue := range searchResp.Issues {
			labels := issue.Fields.Labels
			if labels == nil {
				labels = []string{}
			}
			allResults = append(allResults, SearchResult{
				Key:     issue.Key,
				Summary: issue.Fields.Summary,
				Status:  issue.Fields.Status.Name,
				Labels:  labels,
			})
		}

		if searchResp.NextPageToken == "" {
			break
		}
		nextPageToken = searchResp.NextPageToken
	}

	return allResults, nil
}

func PrintSearchResults(results []SearchResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
