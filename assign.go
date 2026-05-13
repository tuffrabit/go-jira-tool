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

import "fmt"

// AssignIssue assigns the given issue to the currently authenticated user.
func AssignIssue(client *JiraClient, issueKey string) error {
	// Get current user's accountId
	resp, err := client.Get("/myself")
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	var user struct {
		AccountID string `json:"accountId"`
	}
	if err := ReadJSONBody(resp, &user); err != nil {
		return fmt.Errorf("failed to parse current user: %w", err)
	}
	if user.AccountID == "" {
		return fmt.Errorf("current user has no accountId")
	}

	// Assign the ticket
	body := map[string]string{"accountId": user.AccountID}
	resp, err = client.Put("/issue/"+issueKey+"/assignee", body)
	if err != nil {
		return fmt.Errorf("failed to assign issue: %w", err)
	}
	if err := ReadJSONBody(resp, nil); err != nil {
		return fmt.Errorf("assign failed: %w", err)
	}

	return nil
}
