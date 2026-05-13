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
	"fmt"
	"strings"
)

// transitionsResponse represents the response from GET /rest/api/3/issue/{key}/transitions.
type transitionsResponse struct {
	Transitions []transition `json:"transitions"`
}

type transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TransitionIssue transitions a Jira issue to the given status name.
// It fetches available transitions, matches by name (case-insensitive), and posts the transition.
func TransitionIssue(client *JiraClient, issueKey string, targetStatus string) error {
	// GET available transitions
	resp, err := client.Get("/issue/" + issueKey + "/transitions")
	if err != nil {
		return fmt.Errorf("failed to fetch transitions for %s: %w", issueKey, err)
	}

	var result transitionsResponse
	if err := ReadJSONBody(resp, &result); err != nil {
		return fmt.Errorf("failed to read transitions response: %w", err)
	}

	// Match by name (case-insensitive)
	var matched *transition
	var available []string
	for i, t := range result.Transitions {
		available = append(available, t.Name)
		if strings.EqualFold(t.Name, targetStatus) {
			matched = &result.Transitions[i]
		}
	}

	if matched == nil {
		return fmt.Errorf("transition %q not available for %s (available: %s)", targetStatus, issueKey, strings.Join(available, ", "))
	}

	// POST the transition
	body := map[string]interface{}{
		"transition": map[string]string{
			"id": matched.ID,
		},
	}

	resp, err = client.Post("/issue/"+issueKey+"/transitions", body)
	if err != nil {
		return fmt.Errorf("failed to transition %s: %w", issueKey, err)
	}

	if err := ReadJSONBody(resp, nil); err != nil {
		return fmt.Errorf("failed to transition %s: %w", issueKey, err)
	}

	return nil
}
