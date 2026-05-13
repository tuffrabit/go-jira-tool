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
	"os"
)

// AddComment adds a plain-text comment to a Jira issue.
// The body is converted to ADF format before posting.
func AddComment(client *JiraClient, issueKey string, body string) error {
	adfBody := TextToADF(body)

	payload := map[string]json.RawMessage{
		"body": adfBody,
	}

	resp, err := client.Post("/issue/"+issueKey+"/comment", payload)
	if err != nil {
		return fmt.Errorf("failed to add comment to %s: %w", issueKey, err)
	}

	if err := ReadJSONBody(resp, nil); err != nil {
		return fmt.Errorf("failed to add comment to %s: %w", issueKey, err)
	}

	return nil
}

// AddCommentFromFile reads a file and adds its contents as a comment to a Jira issue.
func AddCommentFromFile(client *JiraClient, issueKey string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read comment file: %w", err)
	}
	return AddComment(client, issueKey, string(data))
}
